package entry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pbSync "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
)

func New(
	crypt Crypt,
	client pbSync.SyncServiceClient,
	blobRepo blob.Repository,
) (Service, error) {
	return &service{
		crypt:     crypt,
		client:    client,
		blobRepo:  blobRepo,
		chunkSize: 1024 * 1024, // 1MB, todo get from config
	}, nil
}

type service struct {
	crypt     Crypt
	client    pbSync.SyncServiceClient
	blobRepo  blob.Repository
	chunkSize int64
}

func (s *service) Set(ctx context.Context, key string, name string, entry Entry) error {
	err := s.encryptBlob(entry, key)
	if err != nil {
		return fmt.Errorf("error encrypting entry: %w", err)
	}

	var notes []byte
	if entry.Notes() != "" {
		notes, err = s.encryptNotes(entry.Notes())
		if err != nil {
			return fmt.Errorf("error encrypting notes: %w", err)
		}
	}

	reader, ok, err := s.blobRepo.Reader(key)
	if err != nil {
		return fmt.Errorf("error initializing reader for the encrypted blob: %w", err)
	}
	if !ok {
		return errors.New("encrypted blob that we've just written wasn't found")
	}
	defer reader.Close()

	stream, err := s.client.UpdateEntry(ctx)
	if err != nil {
		return fmt.Errorf("cant start streaming encrypted blob to server: %w", err)
	}
	defer stream.CloseSend()

	// send metadata first
	err = stream.Send(&pbSync.Entry{
		Key:   key,
		Name:  name,
		Notes: notes,
	})
	if err != nil {
		return fmt.Errorf("error sending metadata to the server: %w", err)
	}

	err = s.uploadBlob(reader, stream)
	if err != nil {
		return fmt.Errorf("error uploading encrypted blob to server: %w", err)
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("error finishing upload process: %w", err)
	}

	return nil
}

func (s *service) encryptBlob(entry Entry, key string) (err error) {
	bytesReader, err := entry.Bytes()
	if err != nil {
		return fmt.Errorf("cant get entry bytes reader: %w", err)
	}
	defer func() {
		closeErr := bytesReader.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing entry bytes reader: %w", err)
		}
	}()

	dst, err := s.blobRepo.Writer(key)
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

	// todo cancel on ctx since encrypting can take a while
	_, err = io.Copy(encrypter, bytesReader)
	if err != nil {
		return fmt.Errorf("error writing to encrypted blob: %w", err)
	}

	// but note that defers can also return errors
	return nil
}

func (s *service) uploadBlob(
	blob io.Reader,
	stream grpc.ClientStreamingClient[pbSync.Entry, emptypb.Empty],
) error {
	buffer := make([]byte, s.chunkSize)

	for {
		n, err := blob.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading encrypted blob chunk: %w", err)
		}

		err = stream.Send(&pbSync.Entry{
			Content: buffer[:n],
		})
		if err != nil {
			return fmt.Errorf("error sending encrypted blob chunk to server: %w", err)
		}
	}

	return nil
}

func (s *service) encryptNotes(notes string) ([]byte, error) {
	var buf bytes.Buffer
	encryptWriter, err := s.crypt.Encrypt(&buf)
	if err != nil {
		return nil, fmt.Errorf("could not create encrypt writer: %w", err)
	}

	_, err = io.WriteString(encryptWriter, notes)
	if err != nil {
		_ = encryptWriter.Close()

		return nil, fmt.Errorf("could not write notes to encrypt writer: %w", err)
	}

	err = encryptWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("could not close encrypt writer: %w", err)
	}

	return buf.Bytes(), nil
}

func (s *service) Get(ctx context.Context, key string, entry Entry) (bool, error) {
	stream, err := s.client.GetEntry(ctx, &pbSync.GetEntryRequest{Key: key})
	if err != nil {
		return false, fmt.Errorf("cant start downloading entry: %w", err)
	}
	defer stream.CloseSend()

	resp, err := stream.Recv()
	if err != nil {
		return false, fmt.Errorf("error getting metadata: %w", err)
	}

	if resp.Notes != nil {
		notes, err := s.decryptNotes(resp.Notes)
		if err != nil {
			return false, fmt.Errorf("error decrypting notes: %w", err)
		}
		err = entry.SetNotes(notes)
		if err != nil {
			return false, fmt.Errorf("error setting notes: %w", err)
		}
	}

	content, err := s.downloadBlob(key, stream)
	if err != nil {
		return false, fmt.Errorf("error downloading entry: %w", err)
	}

	decr, err := s.crypt.Decrypt(content)
	if err != nil {
		return false, fmt.Errorf("could not create decrypt reader: %w", err)
	}

	err = entry.FromBytes(decr)
	if err != nil {
		return false, fmt.Errorf("could not read decrypted entry: %w", err)
	}

	return true, nil
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

func (s *service) downloadBlob(key string, stream grpc.ServerStreamingClient[pbSync.Entry]) (io.ReadCloser, error) {
	dst, err := s.blobRepo.Writer(key)
	if err != nil {
		return nil, fmt.Errorf("cant create blob to temporary store entry: %w", err)
	}
	defer dst.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading entry: %w", err)
		}

		_, err = dst.Write(response.Content)
	}

	err = dst.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing temporary blob: %w", err)
	}

	reader, exists, err := s.blobRepo.Reader(key)
	if err != nil {
		return nil, fmt.Errorf("error initializing reader for the encrypted blob: %w", err)
	}
	if !exists {
		return nil, errors.New("encrypted blob that we've just written wasn't found")
	}

	return reader, nil
}

func (s *service) Delete(ctx context.Context, name string) error {
	_, err := s.client.DeleteEntry(ctx, &pbSync.DeleteEntryRequest{Key: name})
	if err != nil {
		return fmt.Errorf("cant delete entry: %w", err)
	}

	return nil
}
