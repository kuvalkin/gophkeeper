package entry

import (
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuvalkin/gophkeeper/internal/client/service"
	pbSync "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
)

func New(
	crypt service.Crypt,
	client pbSync.SyncServiceClient,
	metaRepo MetadataRepository,
	blobRepo BlobRepository,
) (*Service, error) {
	return &Service{
		crypt:     crypt,
		client:    client,
		metaRepo:  metaRepo,
		blobRepo:  blobRepo,
		chunkSize: 1024 * 1024, // 1MB, todo get from config
	}, nil
}

type Service struct {
	crypt     service.Crypt
	client    pbSync.SyncServiceClient
	metaRepo  MetadataRepository
	blobRepo  BlobRepository
	chunkSize int64
}

func (s *Service) Set(ctx context.Context, key string, name string, entry Entry, onConflict func(errMsg string) bool) error {
	size, err := s.saveEncryptedBlob(entry, key)
	if err != nil {
		return fmt.Errorf("error encrypting entry and saving it locally: %w", err)
	}

	lastKnownVersion, ok, err := s.metaRepo.GetVersion(ctx, nil, key)
	if err != nil {
		return fmt.Errorf("error getting last known entry version from local db: %w", err)
	}
	if !ok {
		lastKnownVersion = 0
	}

	reader, ok, err := s.blobRepo.Reader(key)
	if err != nil {
		return fmt.Errorf("error initializing reader for the encrypted blob: %w", err)
	}
	if !ok {
		return errors.New("encrypted blob that we've just written wasn't found")
	}
	defer reader.Close()

	//todo set auth
	stream, err := s.client.UpdateEntry(ctx)
	if err != nil {
		return fmt.Errorf("cant start streaming encrypted blob to server: %w", err)
	}
	defer stream.CloseSend()

	// send metadata first
	err = stream.Send(&pbSync.UpdateEntryRequest{
		Key:         key,
		LastVersion: lastKnownVersion,
	})
	if err != nil {
		err = s.onMetadataSendError(err, onConflict)
		if err != nil {
			return err
		}
	}

	err = s.uploadBlob(reader, stream, size)
	if err != nil {
		return fmt.Errorf("error uploading encrypted blob to server: %w", err)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("error finishing upload process: %w", err)
	}

	// todo transaction?
	err = s.metaRepo.Set(ctx, nil, key, name, response.NewVersion)
	if err != nil {
		return fmt.Errorf("error saving localy uploaded entry metadata: %w", err)
	}

	return nil
}

func (s *Service) saveEncryptedBlob(entry Entry, key string) (size int64, err error) {
	bytesReader, err := entry.Bytes()
	if err != nil {
		return 0, fmt.Errorf("cant get entry bytes reader: %w", err)
	}
	defer func() {
		closeErr := bytesReader.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing entry bytes reader: %w", err)
		}
	}()

	dst, err := s.blobRepo.Writer(key)
	if err != nil {
		return 0, fmt.Errorf("cant create blob to store entry: %w", err)
	}
	defer func() {
		closeErr := dst.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing blob: %w", err)
		}
	}()

	encrypter, err := s.crypt.Encrypt(dst)
	if err != nil {
		return 0, fmt.Errorf("cant start encrypting entry: %w", err)
	}
	defer func() {
		closeErr := encrypter.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("error closing encrypter: %w", err)
		}
	}()

	// todo cancel on ctx since encrypting can take a while
	size, err = io.Copy(encrypter, bytesReader)
	if err != nil {
		return 0, fmt.Errorf("error writing to encrypted blob: %w", err)
	}

	// but note that defers can also return errors
	return size, nil
}

func (s *Service) onMetadataSendError(err error, onConflict func(errMsg string) bool) error {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.AlreadyExists {
		return fmt.Errorf("error sending metadata to the server: %w", err)
	}

	if onConflict != nil && onConflict(st.Message()) {
		// we've decided to continue
		return nil
	}

	return fmt.Errorf("server said that entry already exists: %w", err)
}

func (s *Service) uploadBlob(
	blob io.Reader,
	stream grpc.ClientStreamingClient[pbSync.UpdateEntryRequest, pbSync.UpdateEntryResponse],
	size int64,
) error {
	buffer := make([]byte, s.chunkSize)
	bytesSent := 0

	// todo event upload start

	for {
		// Read next chunk from file
		n, err := blob.Read(buffer)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading encrypted blob chunk: %w", err)
		}

		// Send chunk
		err = stream.Send(&pbSync.UpdateEntryRequest{
			Content: buffer,
		})
		if err != nil {
			return fmt.Errorf("error sending encrypted blob chunk to server: %w", err)
		}

		bytesSent += n
		// todo event upload progress bytesSent out of size
	}

	// todo event upload finish

	return nil
}

func (s *Service) Get(ctx context.Context, name string, entry Entry) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) Delete(ctx context.Context, name string) error {
	//TODO implement me
	panic("implement me")
}
