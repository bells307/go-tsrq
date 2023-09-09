#!/bin/sh

PROTO_PATH=./internal/transport/grpc
protoc --go_out=$PROTO_PATH --go-grpc_out=$PROTO_PATH ./protobuf/tsrq.proto