package sync

import (
	"errors"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
	"github.com/kuvalkin/gophkeeper/internal/server/service/sync"
	"github.com/kuvalkin/gophkeeper/internal/server/transport/auth"
	"github.com/kuvalkin/gophkeeper/internal/support/log"
)

func New(service sync.Service) *Server {
	return &Server{
		service: service,
		log:     log.Logger().Named("server.sync"),
	}
}

type Server struct {
	pb.UnimplementedSyncServiceServer
	service sync.Service
	log     *zap.SugaredLogger
}

func (s *Server) GetEntry(request *pb.GetEntryRequest, stream grpc.ServerStreamingServer[pb.Entry]) error {
	tokenInfo, ok := auth.GetTokenInfo(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no token info")
	}

	llog := s.log.WithLazy("userID", tokenInfo.UserID, "method", "GetEntry", "key", request.Key)

	md, reader, ok, err := s.service.Get(stream.Context(), tokenInfo.UserID, request.Key)
	if err != nil {
		llog.Errorw("cant get entry", "err", err)

		return status.Errorf(codes.Internal, "cant get entry: %v", err)
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

	defer reader.Close()

	buf := make([]byte, 1024*1024) // todo from config
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

func (s *Server) UpdateEntry(stream grpc.ClientStreamingServer[pb.Entry, emptypb.Empty]) error {
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
	uploadChan, resultChan, err := s.service.UpdateEntry(stream.Context(), tokenInfo.UserID, sync.Metadata{
		Key:   request.Key,
		Name:  request.Name,
		Notes: request.Notes,
	})
	if err != nil {
		// todo handle different error types and provide aduquate status
		return status.Errorf(codes.Internal, "cant update entry: %v", err)
	}

	llog = llog.WithLazy("key", request.Key)

	llog.Debug("metadata received, preparing to receive chunks")

	isUploadClosed := false
	defer func() {
		if !isUploadClosed {
			close(uploadChan)
			isUploadClosed = true
		}
	}()

	isDone := false

	for {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "client closed connection")

		case result := <-resultChan:
			if result.Err != nil {
				// todo handle different error types and provide aduquate status
				return status.Errorf(codes.Internal, "cant update entry: %v", err)
			}

			return stream.SendAndClose(&emptypb.Empty{})

		default:
			if isDone {
				llog.Debug("waiting for result")

				// todo read steam from other goroutine, pipe in chan
				time.Sleep(10 * time.Millisecond)
				continue
			}

			llog.Debug("waiting for chunk")

			req, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				llog.Debug("uploaded ended, wait for result")

				if !isUploadClosed {
					close(uploadChan)
					isUploadClosed = true
				}

				isDone = true

				continue
			}

			if err != nil {
				uploadChan <- sync.UploadChunk{Err: fmt.Errorf("cant get chunk: %w", err)}

				if !isUploadClosed {
					close(uploadChan)
					isUploadClosed = true
				}

				isDone = true

				continue
			}

			uploadChan <- sync.UploadChunk{Content: req.Content}
			llog.Debug("chunk uploaded")
		}
	}
}

func (s *Server) DeleteEntry(stream grpc.ClientStreamingServer[pb.DeleteEntryRequest, emptypb.Empty]) error {
	panic("implement me")
}
