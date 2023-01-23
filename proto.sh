#!/bin/bash

swagger="static/swagger"
[ ! -d "$swagger" ] && mkdir -p "$swagger"

for d in server/proto/* ; do

    [ -L "${d%/}" ] && continue
    [ "${d%/}" == "server/proto/apis" ] && continue

    protoc -I=. -I="$(cd "$(dirname -- "$1")" >/dev/null; pwd -P)/$(basename -- "$1")server/proto/apis" --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. "$d"/*.proto
    protoc -I=. -I="$(cd "$(dirname -- "$1")" >/dev/null; pwd -P)/$(basename -- "$1")server/proto/apis" --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. --swagger_out=logtostderr=true:./static/swagger "$d"/*.proto

    echo "$d - BUILD SUCCESS!"
done

for d in static/swagger/server/proto/* ; do
    [ -L "${d%/}" ] && continue
    cp -r "$d"/* static/swagger
done

rm -r static/swagger/server