#!/bin/sh

grpcurl -plaintext -d '{"id":"123","data":"somedata"}' localhost:9999 grpc.TSRQService/Enqueue
grpcurl -plaintext -d '{"id":"456","data":"somedata"}' localhost:9999 grpc.TSRQService/Enqueue
grpcurl -plaintext -d '{"id":"789","data":"somedata"}' localhost:9999 grpc.TSRQService/Enqueue