// Package entry provides the gRPC server implementation for managing entries.
// It includes methods for retrieving, setting, and deleting entries, as well as handling
// streaming data for large content uploads and downloads.
package entry

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/kuvalkin/gophkeeper/internal/server/service/entry"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/auth"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
	"github.com/kuvalkin/gophkeeper/internal/support/utils"
	pb "github.com/kuvalkin/gophkeeper/pkg/proto/entry/v1"
)

// New creates a new instance of the EntryServiceServer.
// It takes an entry.Service implementation and a chunk size for streaming data.
// Returns an implementation of pb.EntryServiceServer.
func New(service entry.Service, chunkSize int64) pb.EntryServiceServer {
	return &server{
		service:   service,
		chunkSize: chunkSize,
		log:       log.Logger().Named("server.entry"),
	}
}

type server struct {
	pb.UnsafeEntryServiceServer
	service   entry.Service
	chunkSize int64 // Size of chunks for streaming data.
	log       *zap.SugaredLogger
}

// GetEntry streams an entry's metadata and content to the client.
// It retrieves the entry by key and streams the data in chunks.
// Returns an error if the operation fails.
func (s *server) GetEntry(request *pb.GetEntryRequest, stream grpc.ServerStreamingServer[pb.Entry]) error {
	tokenInfo, ok := auth.GetTokenInfo(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no token info")
	}

	llog := s.log.WithLazy("userID", tokenInfo.UserID, "method", "GetEntry", "key", request.Key)

	md, reader, ok, err := s.service.GetEntry(stream.Context(), tokenInfo.UserID, request.Key)
	if err != nil {
		return status.Errorf(codes.Internal, "cant get entry")
	}
	if !ok {
		return status.Errorf(codes.NotFound, "entry not found")
	}

	err = stream.Send(&pb.Entry{
		Key:   md.Key,
		Name:  md.Name,
		Notes: md.Notes,
	})
	if err != nil {
		llog.Errorw("cant send metadata", "err", err)

		return status.Error(codes.Internal, "cant send metadata")
	}

	defer utils.CloseAndLogError(reader, llog)

	buf := make([]byte, s.chunkSize)
	for {
		n, err := reader.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			llog.Errorw("cant read entry", "err", err)

			return status.Error(codes.Internal, "cant read entry")
		}

		err = stream.Send(&pb.Entry{
			Content: buf[:n],
		})
		if err != nil {
			llog.Errorw("cant send entry", "err", err)

			return status.Error(codes.Internal, "cant send entry")
		}
	}

	return nil
}

// SetEntry handles bidirectional streaming for uploading an entry's content.
// It receives metadata and content chunks from the client and stores them.
// Returns an error if the operation fails.
func (s *server) SetEntry(stream grpc.BidiStreamingServer[pb.SetEntryRequest, pb.SetEntryResponse]) error {
	tokenInfo, ok := auth.GetTokenInfo(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no token info")
	}

	llog := s.log.WithLazy("userID", tokenInfo.UserID, "method", "UpdateEntry")

	// get the metadata
	request, err := stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return status.Error(codes.Canceled, "client closed connection")
		}

		llog.Errorw("cant get metadata", "err", err)

		return status.Error(codes.Internal, "cant get metadata")
	}
	if request.Entry == nil || request.Entry.Key == "" || request.Entry.Name == "" {
		return status.Error(codes.InvalidArgument, "metadata is empty")
	}

	uploadChan, resultChan, err := s.service.SetEntry(stream.Context(), tokenInfo.UserID, entry.Metadata{
		Key:   request.Entry.Key,
		Name:  request.Entry.Name,
		Notes: request.Entry.Notes,
	}, request.Overwrite)
	if err != nil && !errors.Is(err, entry.ErrEntryExists) {
		return status.Error(codes.Internal, "cant set entry")
	}

	llog = llog.WithLazy("key", request.Entry.Key)

	if errors.Is(err, entry.ErrEntryExists) {
		err = stream.Send(&pb.SetEntryResponse{AlreadyExists: true})
		if err != nil {
			if errors.Is(err, io.EOF) {
				return status.Error(codes.Canceled, "client closed connection")
			}

			llog.Errorw("cant send already exists", "err", err)

			return status.Error(codes.Internal, "cant send already exists")
		}

		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return status.Error(codes.Canceled, "client closed connection")
			}

			llog.Errorw("cant get already exists response", "err", err)

			return status.Error(codes.InvalidArgument, "cant get already exists response")
		}

		if !resp.Overwrite {
			return status.Errorf(codes.AlreadyExists, "entry already exists")
		}

		llog.Debug("client sent overwrite signal")

		// continue and overwrite
		uploadChan, resultChan, err = s.service.SetEntry(stream.Context(), tokenInfo.UserID, entry.Metadata{
			Key:   request.Entry.Key,
			Name:  request.Entry.Name,
			Notes: request.Entry.Notes,
		}, true)
		if err != nil {
			llog.Errorw("cant set entry with overwrite", "err", err)

			return status.Error(codes.Internal, "cant set entry")
		}
	} else {
		// send client signal that it can start uploading
		err = stream.Send(&pb.SetEntryResponse{})
		if err != nil {
			if errors.Is(err, io.EOF) {
				return status.Error(codes.Canceled, "client closed connection")
			}

			llog.Errorw("cant send start upload signal", "err", err)

			return status.Error(codes.Internal, "cant send start upload signal")
		}
	}

	llog.Debug("metadata received, preparing to receive chunks")

	go s.downloadContentChunks(stream.Context(), stream, uploadChan, llog.Named("upload"))

	select {
	case <-stream.Context().Done():
		return status.Error(codes.Canceled, "client closed connection")

	case result := <-resultChan:
		if result.Err != nil {
			if errors.Is(err, context.Canceled) {
				return status.Error(codes.Canceled, "client closed connection")
			}

			if errors.Is(err, entry.ErrUploadChunk) {
				return status.Error(codes.Internal, "cant upload chunk")
			}

			return status.Error(codes.Internal, "internal error during upload")
		}

		return nil
	}
}

// DeleteEntry deletes an entry identified by its key.
// It requires authentication and returns an error if the operation fails.
func (s *server) DeleteEntry(ctx context.Context, request *pb.DeleteEntryRequest) (*emptypb.Empty, error) {
	tokenInfo, ok := auth.GetTokenInfo(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no token info")
	}

	err := s.service.DeleteEntry(ctx, tokenInfo.UserID, request.Key)
	if err != nil {
		return nil, status.Error(codes.Internal, "cant delete entry")
	}

	return &emptypb.Empty{}, nil
}

// downloadContentChunks reads content chunks from the client and sends them to the upload channel.
func (s *server) downloadContentChunks(ctx context.Context, stream grpc.BidiStreamingServer[pb.SetEntryRequest, pb.SetEntryResponse], uploadChan chan<- entry.UploadChunk, llog *zap.SugaredLogger) {
	defer close(uploadChan)

	for {
		select {
		case <-ctx.Done():
			llog.Debug("context done")
			return
		default:
			llog.Debug("waiting for chunk")

			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				llog.Debug("uploaded ended")

				return
			}
			if stErr, ok := status.FromError(err); ok && stErr.Code() == codes.Canceled {
				llog.Debug("client closed connection")

				return
			}

			if err != nil {
				llog.Errorw("cant get chunk", "err", err)

				uploadChan <- entry.UploadChunk{Err: fmt.Errorf("cant get chunk: %w", err)}

				return
			}

			llog.Debug("uploading chunk")
			uploadChan <- entry.UploadChunk{Content: req.Entry.Content}
			llog.Debug("chunk uploaded")
		}
	}
}
