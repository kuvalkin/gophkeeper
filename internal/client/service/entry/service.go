package entry

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kuvalkin/gophkeeper/internal/client/cmd"
	pbSync "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
	"github.com/kuvalkin/gophkeeper/internal/storage/blob"
)

func New(
	crypt Crypt,
	client pbSync.SyncServiceClient,
	blobRepo blob.Repository,
) (*Service, error) {
	return &Service{
		crypt:     crypt,
		client:    client,
		blobRepo:  blobRepo,
		chunkSize: 1024 * 1024, // 1MB, todo get from config
	}, nil
}

type Service struct {
	crypt     Crypt
	client    pbSync.SyncServiceClient
	blobRepo  blob.Repository
	chunkSize int64
}

func (s *Service) Set(ctx context.Context, key string, name string, entry cmd.Entry) error {
	err := s.encryptBlob(entry, key)
	if err != nil {
		return fmt.Errorf("error encrypting entry: %w", err)
	}

	notes, err := s.encryptNotes(entry.Notes())
	if err != nil {
		return fmt.Errorf("error encrypting notes: %w", err)
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

func (s *Service) encryptBlob(entry cmd.Entry, key string) (err error) {
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

func (s *Service) uploadBlob(
	blob io.Reader,
	stream grpc.ClientStreamingClient[pbSync.Entry, emptypb.Empty],
) error {
	buffer := make([]byte, s.chunkSize)
	bytesSent := 0

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
		err = stream.Send(&pbSync.Entry{
			Content: buffer,
		})
		if err != nil {
			return fmt.Errorf("error sending encrypted blob chunk to server: %w", err)
		}

		bytesSent += n
	}

	return nil
}

func (s *Service) encryptNotes(notes string) ([]byte, error) {
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

func (s *Service) Get(ctx context.Context, name string, entry cmd.Entry) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Service) Delete(ctx context.Context, name string) error {
	//TODO implement me
	panic("implement me")
}
