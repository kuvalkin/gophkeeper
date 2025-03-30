package entry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
	pb "github.com/kuvalkin/gophkeeper/pkg/proto/entry/v1"
)

// New creates a new instance of the entry service, which encapsulates the core business logic
// for handling entries. This includes encryption, decryption, and interaction with the server
// and local storage for managing entry data.
func New(
	crypt Crypt,
	client pb.EntryServiceClient,
	blobRepo blob.Repository,
	chunkSize int64,
) Service {
	return &service{
		crypt:     crypt,
		client:    client,
		blobRepo:  blobRepo,
		chunkSize: chunkSize,
		log:       log.Logger().Named("service.entry"),
	}
}

type service struct {
	crypt     Crypt
	client    pb.EntryServiceClient
	blobRepo  blob.Repository
	chunkSize int64
	log       *zap.SugaredLogger
}

// SetEntry creates or updates an entry with the given key, name, notes, and content.
// It encrypts the content and uploads it to the server. If the entry already exists,
// the onOverwrite callback determines whether to overwrite it.
func (s *service) SetEntry(ctx context.Context, key string, name string, notes string, content io.ReadCloser, onOverwrite func() bool) error {
	llog := s.log.WithLazy("key", key, "name", name)

	llog.Debug("encrypting entry content")
	err := s.encryptBlob(ctx, content, key)
	if err != nil {
		return fmt.Errorf("error encrypting entry: %w", err)
	}

	var encNotes []byte
	if notes != "" {
		llog.Debug("encrypting notes")
		encNotes, err = s.encryptNotes(notes)
		if err != nil {
			return fmt.Errorf("error encrypting notes: %w", err)
		}
	}

	reader, ok, err := s.blobRepo.OpenBlobReader(key)
	if err != nil {
		return fmt.Errorf("error initializing reader for the encrypted blob: %w", err)
	}
	if !ok {
		return errors.New("encrypted blob that we've just written wasn't found")
	}
	defer utils.CloseAndLogError(reader, llog)

	llog.Debug("opening grpc steam")
	stream, err := s.client.SetEntry(ctx)
	if err != nil {
		return fmt.Errorf("cant start streaming encrypted blob to server: %w", err)
	}
	defer func() {
		err = stream.CloseSend()
		if err != nil {
			llog.Errorw("error closing stream", "err", err)
		}
	}()

	llog.Debug("sending metadata")
	err = s.sendMetadata(stream, &pb.Entry{
		Key:   key,
		Name:  name,
		Notes: encNotes,
	}, onOverwrite)
	if err != nil {
		return fmt.Errorf("error sending metadata to the server: %w", err)
	}

	llog.Debug("uploading encrypted content")
	err = s.uploadBlob(ctx, reader, stream)
	if err != nil {
		return fmt.Errorf("error uploading encrypted blob to server: %w", err)
	}

	llog.Debug("upload done")
	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("error finishing upload process: %w", err)
	}

	llog.Debug("getting server acknowledgement")
	_, err = stream.Recv()
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("error getting acknowledgement from server: %w", err)
	}

	return nil
}

func (s *service) encryptNotes(notes string) (b []byte, err error) {
	var buf bytes.Buffer
	encryptWriter, err := s.crypt.Encrypt(&buf)
	if err != nil {
		return nil, fmt.Errorf("could not create encrypt writer: %w", err)
	}
	defer func() {
		closeErr := encryptWriter.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing encrypt writer: %w", err)
		}
	}()

	_, err = io.WriteString(encryptWriter, notes)
	if err != nil {
		return nil, fmt.Errorf("could not write notes to encrypt writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *service) encryptBlob(ctx context.Context, content io.ReadCloser, key string) (err error) {
	defer func() {
		closeErr := content.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing entry bytes reader: %w", err)
		}
	}()

	dst, err := s.blobRepo.OpenBlobWriter(key)
	if err != nil {
		return fmt.Errorf("cant create blob to store entry: %w", err)
	}
	defer func() {
		closeErr := dst.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing blob: %w", err)
		}
	}()

	encrypter, err := s.crypt.Encrypt(dst)
	if err != nil {
		return fmt.Errorf("cant start encrypting entry: %w", err)
	}
	defer func() {
		closeErr := encrypter.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing encrypter: %w", err)
		}
	}()

	_, err = utils.CopyContext(ctx, encrypter, content)
	if err != nil {
		return fmt.Errorf("error writing to encrypted blob: %w", err)
	}

	// but note that defers can also return errors
	return nil
}

func (s *service) sendMetadata(stream grpc.BidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse], entry *pb.Entry, onOverwrite func() bool) error {
	err := stream.Send(&pb.SetEntryRequest{
		Entry: entry,
	})
	if err != nil {
		return fmt.Errorf("error sending initial request: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.AlreadyExists {
			// errors means that stream is closed, no sense in continuing
			return ErrEntryExists
		}

		return fmt.Errorf("error receiving response: %w", err)
	}

	if !resp.AlreadyExists {
		return nil
	}

	if onOverwrite == nil || !onOverwrite() {
		return ErrEntryExists
	}

	err = stream.Send(&pb.SetEntryRequest{
		Overwrite: true,
	})
	if err != nil {
		return fmt.Errorf("error sending overwrite signal to the server: %w", err)
	}

	return nil
}

func (s *service) uploadBlob(
	ctx context.Context,
	blob io.Reader,
	stream grpc.BidiStreamingClient[pb.SetEntryRequest, pb.SetEntryResponse],
) error {
	buffer := make([]byte, s.chunkSize)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("uploading was interrupted: %w", ctx.Err())
		default:
			n, err := blob.Read(buffer)

			if errors.Is(err, io.EOF) {
				return nil
			}

			if err != nil {
				return fmt.Errorf("error reading encrypted blob chunk: %w", err)
			}

			err = stream.Send(&pb.SetEntryRequest{
				Entry: &pb.Entry{
					Content: buffer[:n],
				},
			})
			if err != nil {
				return fmt.Errorf("error sending encrypted blob chunk to server: %w", err)
			}
		}
	}
}

// GetEntry retrieves an entry by its key. It decrypts the notes and content and returns them.
// The boolean indicates whether the entry exists, and an error is returned if any issues occur.
func (s *service) GetEntry(ctx context.Context, key string) (string, io.ReadCloser, bool, error) {
	stream, err := s.client.GetEntry(ctx, &pb.GetEntryRequest{Key: key})
	if err != nil {
		return "", nil, false, fmt.Errorf("cant start downloading entry: %w", err)
	}
	defer func() {
		err = stream.CloseSend()
		if err != nil {
			s.log.Errorw("error closing stream", "err", err)
		}
	}()

	resp, err := stream.Recv()
	if err != nil {
		if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.NotFound {
			return "", nil, false, nil
		}

		return "", nil, false, fmt.Errorf("error getting metadata: %w", err)
	}

	var notes string
	if resp.Notes != nil {
		notes, err = s.decryptNotes(resp.Notes)
		if err != nil {
			return "", nil, false, fmt.Errorf("error decrypting notes: %w", err)
		}
	}

	content, err := s.downloadBlob(key, stream)
	if err != nil {
		return "", nil, false, fmt.Errorf("error downloading entry: %w", err)
	}

	decr, err := s.crypt.Decrypt(content)
	if err != nil {
		return "", nil, false, fmt.Errorf("could not create decrypt reader: %w", err)
	}

	return notes, &combinedRC{reader: decr, closer: content}, true, nil
}

func (s *service) decryptNotes(encNotes []byte) (string, error) {
	dec, err := s.crypt.Decrypt(bytes.NewReader(encNotes))
	if err != nil {
		return "", fmt.Errorf("could not create decrypt reader: %w", err)
	}

	notes, err := io.ReadAll(dec)
	if err != nil {
		return "", fmt.Errorf("could not read decrypted notes: %w", err)
	}

	return string(notes), nil
}

func (s *service) downloadBlob(key string, stream grpc.ServerStreamingClient[pb.Entry]) (io.ReadCloser, error) {
	dst, err := s.blobRepo.OpenBlobWriter(key)
	if err != nil {
		return nil, fmt.Errorf("cant create blob to temporary store entry: %w", err)
	}
	defer utils.CloseAndLogError(dst, s.log)

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading entry: %w", err)
		}

		_, err = dst.Write(response.Content)
		if err != nil {
			return nil, fmt.Errorf("error writing entry: %w", err)
		}
	}

	err = dst.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing temporary blob: %w", err)
	}

	reader, exists, err := s.blobRepo.OpenBlobReader(key)
	if err != nil {
		return nil, fmt.Errorf("error initializing reader for the encrypted blob: %w", err)
	}
	if !exists {
		return nil, errors.New("encrypted blob that we've just written wasn't found")
	}

	return reader, nil
}

// DeleteEntry removes an entry by its key. Returns an error if the deletion fails.
func (s *service) DeleteEntry(ctx context.Context, name string) error {
	_, err := s.client.DeleteEntry(ctx, &pb.DeleteEntryRequest{Key: name})
	if err != nil {
		return fmt.Errorf("cant delete entry: %w", err)
	}

	return nil
}

type combinedRC struct {
	reader io.Reader
	closer io.Closer
}

func (c *combinedRC) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func (c *combinedRC) Close() error {
	return c.closer.Close()
}
