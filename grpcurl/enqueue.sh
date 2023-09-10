#!/bin/sh

grpcurl -plaintext -d '{"id":"456","data":"{\"someField\":\"someText\",\"anotherField\":123}"}' localhost:9999 grpc.TSRQService/Enqueue
