CLIENT_BIN := gkeep
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +'%Y-%m-%d')
LDFLAGS = -X 'github.com/kuvalkin/gophkeeper/internal/client/cmd.version=$(VERSION)' \
          -X 'github.com/kuvalkin/gophkeeper/internal/client/cmd.buildDate=$(BUILD_DATE)'

generate: generate-auth generate-entry generate-serialize go-generate

generate-auth:
	protoc \
		-I api/proto/vendor/protovalidate/proto/protovalidate \
		-I . \
 		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		api/proto/auth/v1/*.proto

generate-entry:
	protoc \
		-I api/proto/vendor/protovalidate/proto/protovalidate \
		-I . \
 		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		api/proto/entry/v1/*.proto

generate-serialize:
	protoc \
		-I api/proto/vendor/protovalidate/proto/protovalidate \
		-I . \
 		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		api/proto/serialize/v1/*.proto

go-generate:
	go generate ./...

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./... -v

test:
	go test ./...

build-client: 
	go mod tidy
	
	rm -rf bin/client/*
	GOOS=linux GOARCH=amd64 go build -v -ldflags "$(LDFLAGS)" -buildvcs=false -o=bin/client/$(CLIENT_BIN)-linux-amd64 ./cmd/client/...
	GOOS=windows GOARCH=amd64 go build -v -ldflags "$(LDFLAGS)" -buildvcs=false -o=bin/client/$(CLIENT_BIN)-windows-amd64.exe ./cmd/client/...
	GOOS=darwin GOARCH=amd64 go build -v -ldflags "$(LDFLAGS)" -buildvcs=false -o=bin/client/$(CLIENT_BIN)-darwin-amd64 ./cmd/client/...
	GOOS=darwin GOARCH=arm64 go build -v -ldflags "$(LDFLAGS)" -buildvcs=false -o=bin/client/$(CLIENT_BIN)-darwin-arm64 ./cmd/client/...
	chmod -R +x bin
