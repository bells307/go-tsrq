#!/bin/sh

grpcurl -plaintext -d '{"id":"1223"}' localhost:9999 grpc.TSRQService/Exists