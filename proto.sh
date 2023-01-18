#!/bin/bash

swagger="static/swagger"
[ ! -d "$swagger" ] && mkdir -p "$swagger"

for d in server/proto/* ; do

    [ -L "${d%/}" ] && continue

    protoc -I=. -I="$GOPATH"/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. "$d"/*.proto
    protoc -I=. -I="$GOPATH"/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. --swagger_out=logtostderr=true:./static/swagger "$d"/*.proto

    echo "$d - BUILD SUCCESS!"
done

for d in static/swagger/server/proto/* ; do
    [ -L "${d%/}" ] && continue
    cp -r "$d"/* static/swagger
done

rm -r static/swagger/server