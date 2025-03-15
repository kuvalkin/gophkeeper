package sync

import (
	"context"

	"google.golang.org/grpc"

	pb "github.com/kuvalkin/gophkeeper/internal/proto/sync/v1"
)

func New() pb.SyncServiceServer {
	return &server{}
}

type server struct {
	pb.UnimplementedSyncServiceServer
}

func (s *server) GetUpdates(request *pb.GetUpdatesRequest, g grpc.ServerStreamingServer[pb.GetUpdatesResponse]) error {
	//TODO implement me
	panic("implement me")
}

func (s *server) UpdateEntry(g grpc.ClientStreamingServer[pb.UpdateEntryRequest, pb.UpdateEntryResponse]) error {
	//TODO implement me
	panic("implement me")
}

func (s *server) DeleteEntry(ctx context.Context, request *pb.DeleteEntryRequest) (*pb.DeleteEntryResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *server) DownloadEntry(request *pb.DownloadEntryRequest, g grpc.ServerStreamingServer[pb.DownloadEntryResponse]) error {
	//TODO implement me
	panic("implement me")
}
