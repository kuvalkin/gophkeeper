package sync

import (
	"context"
	"errors"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
	"github.com/kuvalkin/gophkeeper/internal/server/service/sync"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/auth"
)

func New(service sync.Service) pb.SyncServiceServer {
	return &server{}
}

type server struct {
	pb.UnimplementedSyncServiceServer
	service sync.Service
	log     *zap.SugaredLogger
}

func (s *server) GetUpdates(request *pb.GetUpdatesRequest, g grpc.ServerStreamingServer[pb.GetUpdatesResponse]) error {
	//TODO implement me
	panic("implement me")
}

func (s *server) UpdateEntry(stream grpc.ClientStreamingServer[pb.UpdateEntryRequest, pb.UpdateEntryResponse]) error {
	tokenInfo, ok := auth.GetTokenInfo(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no token info")
	}

	llog := s.log.WithLazy("userID", tokenInfo.UserID, "method", "UpdateEntry")

	// get the metadata
	request, err := stream.Recv()
	if err != nil {
		llog.Errorw("cant get metadata", "err", err)

		return status.Errorf(codes.InvalidArgument, "cant get metadata: %v", err)
	}

	// todo maybe do smth similar with pipes and chans on client?
	writer, err, resultChan := s.service.UpdateEntry(stream.Context(), tokenInfo.UserID, sync.Metadata{
		Key:   request.Key,
		Name:  request.Name,
		Notes: request.Notes,
	}, request.LastVersion, request.Force)
	if err != nil {
		if errors.Is(err, sync.ErrVersionMismatch) {
			return status.Errorf(codes.FailedPrecondition, "version mismatch: %v", err)
		}

		// todo handle different error types and provide aduquate status
		return status.Errorf(codes.Internal, "cant update entry: %v", err)
	}

	go func() {
		defer writer.Close()

		for {
			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}

			if err != nil {
				writer.CloseWithError(err)
				return
			}

			if _, err := writer.Write(req.Content); err != nil {
				writer.CloseWithError(err)
				return
			}
		}
	}()

	result := <-resultChan
	if result.Err != nil {
		// todo handle different error types and provide aduquate status
		return status.Errorf(codes.Internal, "cant update entry: %v", err)
	}

	return stream.SendAndClose(&pb.UpdateEntryResponse{NewVersion: result.NewVersion})
}

func (s *server) DeleteEntry(ctx context.Context, request *pb.DeleteEntryRequest) (*pb.DeleteEntryResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) DownloadEntry(request *pb.DownloadEntryRequest, g grpc.ServerStreamingServer[pb.DownloadEntryResponse]) error {
	//TODO implement me
	panic("implement me")
}
