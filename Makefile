generate: generate-auth generate-entry generate-serialize

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

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./... -v