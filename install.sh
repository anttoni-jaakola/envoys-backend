#!/bin/bash

sudo apt update
sudo apt install -y build-essential protobuf-compiler libprotobuf-dev

go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
go install github.com/golang/protobuf/protoc-gen-go
go install github.com/golang/protobuf/{proto,protoc-gen-go}

exit