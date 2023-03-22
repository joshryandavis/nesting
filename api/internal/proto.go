package internal

//go:generate protoc -I ./ ./proto/nesting.proto --go_out=./
//go:generate protoc -I ./ ./proto/nesting.proto --go-grpc_out=./
//go:generate mockery --with-expecter --srcpkg ./proto --output ./proto/mocks --name=NestingClient
