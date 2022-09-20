package internal

//go:generate protoc -I ./ ./proto/nesting.proto --go_out=./
//go:generate protoc -I ./ ./proto/nesting.proto --go-grpc_out=./
