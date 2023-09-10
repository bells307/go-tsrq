#!/bin/sh

grpcurl -plaintext -d '{"id":"123","data":{"field": 2}}' localhost:9999 grpc.TSRQService/Enqueue
grpcurl -plaintext -d '{"id":"456","data":{"field": "text"}}' localhost:9999 grpc.TSRQService/Enqueue
grpcurl -plaintext -d '{"id":"789","data":{"field": {"moreField":1}}}' localhost:9999 grpc.TSRQService/Enqueue