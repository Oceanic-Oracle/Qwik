.PHONY: proto

proto:
	protoc -I ../proto \
	--go_out=paths=source_relative:internal/server/grpc-server/gen \
	--go-grpc_out=paths=source_relative:internal/server/grpc-server/gen \
	mailer.proto