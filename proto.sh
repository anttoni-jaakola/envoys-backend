#!/bin/bash

grpc_path="$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway"
if ! [ -f $grpc_path ]; then
    grpc_path=$(find $GOPATH/pkg/mod/github.com/grpc-ecosystem -name "grpc-gateway@*" -type d |  head -n 1)
fi

directories=$(echo "$PATH" | tr ":" "\n")
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

for d in $(find server -name '*.proto') ; do

    [ -L "${d%/}" ] && continue

    googleapis="$grpc_path/third_party/googleapis"
    protoc -I=. -I="$googleapis" --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. "$d"
    protoc -I=. -I="$googleapis" --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. --swagger_out=logtostderr=true:. "$d"
    echo "$d - BUILD SUCCESS!"
done

cp -r github.com/cryptogateway/backend-envoys/server/types/* server/types/
rm -r github.com