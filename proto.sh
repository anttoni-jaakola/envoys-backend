#!/bin/bash

swagger="static/swagger"
[ ! -d "$swagger" ] && mkdir -p "$swagger"

grpc_path="$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway"
if ! [ -f $grpc_path ]; then
    grpc_path=$(find $GOPATH/pkg/mod/github.com/grpc-ecosystem -name "grpc-gateway@*" -type d |  head -n 1)
fi

directories=$(echo $PATH | tr ":" "\n")
findedGOPATH="false"
for directory in $directories
do
    if [ "$directory" == "$GOPATH/bin" ]; then
        findedGOPATH="true"
    fi
done

if [ "$findedGOPATH" == "false" ]; then
    echo "add $GOPATH/bin to PATH"
    PATH="$PATH:$GOPATH/bin"
fi

for d in server/proto/* ; do

    [ -L "${d%/}" ] && continue
    googleapis="$grpc_path/third_party/googleapis"
    echo $googleapis
    protoc -I=. -I=$googleapis --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. "$d"/*.proto
    protoc -I=. -I=$googleapis --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. --swagger_out=logtostderr=true:./static/swagger "$d"/*.proto

    echo "$d - BUILD SUCCESS!"
done

for d in static/swagger/server/proto/* ; do
    [ -L "${d%/}" ] && continue
    cp -r "$d"/* static/swagger
done

rm -r static/swagger/server