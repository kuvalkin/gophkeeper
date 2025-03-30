package mocks

//go:generate mockgen -destination=./blob_repository_mock.go -package=mocks -mock_names Repository=MockBlobRepository github.com/kuvalkin/gophkeeper/internal/storage/blob Repository
//go:generate mockgen -destination=./meta_repository_mock.go -package=mocks github.com/kuvalkin/gophkeeper/internal/server/service/entry MetadataRepository
//go:generate mockgen -destination=./write_closer_mock.go -package=mocks io WriteCloser
//go:generate mockgen -destination=./read_closer_mock.go -package=mocks io ReadCloser
//go:generate mockgen -destination=./user_repository_mock.go -package=mocks -mock_names Repository=MockUserRepository github.com/kuvalkin/gophkeeper/internal/server/service/user Repository
//go:generate mockgen -destination=./user_service_mock.go -package=mocks -mock_names Service=MockUserService github.com/kuvalkin/gophkeeper/internal/server/service/user Service
//go:generate mockgen -destination=./entry_service_mock.go -package=mocks -mock_names Service=MockEntryService github.com/kuvalkin/gophkeeper/internal/server/service/entry Service
//go:generate mockgen -destination=./bidi_stream_mock.go -package=mocks google.golang.org/grpc BidiStreamingServer
//go:generate mockgen -destination=./server_stream_mock.go -package=mocks google.golang.org/grpc ServerStreamingServer
