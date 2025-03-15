generate: generate-auth generate-sync generate-serialize

generate-auth:
	protoc \
		-I api/proto/vendor/protovalidate/proto/protovalidate \
		-I . \
 		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		api/proto/auth/v1/*.proto

generate-sync:
	protoc \
		-I api/proto/vendor/protovalidate/proto/protovalidate \
		-I . \
 		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		api/proto/sync/v1/*.proto

generate-serialize:
	protoc \
		-I api/proto/vendor/protovalidate/proto/protovalidate \
		-I . \
 		--go_out=. --go_opt=paths=import \
		--go-grpc_out=. --go-grpc_opt=paths=import \
		api/proto/serialize/v1/*.proto

