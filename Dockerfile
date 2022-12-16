FROM golang:1.18.9 as builder

RUN apt-get update

COPY . "$GOPATH/src/github.com/cryptogateway/backend-envoys"
WORKDIR "$GOPATH/src/github.com/cryptogateway/backend-envoys"

# Install tools
RUN apt-get install -y build-essential curl libkrb5-dev cmake

# Install grpc
RUN apt install -y protobuf-compiler libprotobuf-dev
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN go install github.com/golang/protobuf/proto
RUN go install github.com/golang/protobuf/protoc-gen-go

RUN ./proto.sh
RUN go build main.go
CMD ["./main"]