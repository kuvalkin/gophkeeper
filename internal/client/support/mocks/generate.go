package mocks

//go:generate mockgen -destination=./bidi_stream_mock.go -package=mocks google.golang.org/grpc BidiStreamingClient
//go:generate mockgen -destination=./blob_repository_mock.go -package=mocks -mock_names Repository=MockBlobRepository  github.com/kuvalkin/gophkeeper/internal/storage/blob Repository
//go:generate mockgen -destination=./entry_client_mock.go -package=mocks github.com/kuvalkin/gophkeeper/internal/proto/entry/v1 EntryServiceClient
//go:generate mockgen -destination=./read_closer_mock.go -package=mocks io ReadCloser
//go:generate mockgen -destination=./write_closer_mock.go -package=mocks io WriteCloser
//go:generate mockgen -destination=./server_stream_mock.go -package=mocks google.golang.org/grpc ServerStreamingClient
//go:generate mockgen -destination=./crypt_mock.go -package=mocks github.com/kuvalkin/gophkeeper/internal/client/service/entry Crypt
//go:generate mockgen -destination=./container_mock.go -package=mocks github.com/kuvalkin/gophkeeper/internal/client/service/container Container
//go:generate mockgen -destination=./auth_service_mock.go -package=mocks -mock_names Service=MockAuthService github.com/kuvalkin/gophkeeper/internal/client/service/auth Service
//go:generate mockgen -destination=./entry_service_mock.go -package=mocks -mock_names Service=MockEntryService github.com/kuvalkin/gophkeeper/internal/client/service/entry Service
//go:generate mockgen -destination=./secret_service_mock.go -package=mocks -mock_names Service=MockSecretService github.com/kuvalkin/gophkeeper/internal/client/service/secret Service
//go:generate mockgen -destination=./prompter_mock.go -package=mocks github.com/kuvalkin/gophkeeper/internal/client/tui/prompts Prompter
//go:generate mockgen -destination=./auth_repository_mock.go -package=mocks -mock_names Repository=MockAuthRepository github.com/kuvalkin/gophkeeper/internal/client/service/auth Repository
//go:generate mockgen -destination=./auth_client_mock.go -package=mocks github.com/kuvalkin/gophkeeper/internal/proto/auth/v1 AuthServiceClient
//go:generate mockgen -destination=./secret_repository_mock.go -package=mocks -mock_names Repository=MockSecretRepository github.com/kuvalkin/gophkeeper/internal/client/service/secret Repository
