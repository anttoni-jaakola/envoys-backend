## Install to server
`./install.sh`
****

## Proto build
`sudo apt install protobuf-compiler libprotobuf-dev`  
`go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway`  
`go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger`  
`go install github.com/golang/protobuf/protoc-gen-go`  
`go install github.com/golang/protobuf/{proto,protoc-gen-go}`

`./proto.sh`
****

## Docker
`docker-compose up --build`
****

| Type       | Supported |
|------------|-----------|
| 0 - Spot   | Yes       |
| 1 - Stock  | Dev       |
| 2 - Margin | Dev       |