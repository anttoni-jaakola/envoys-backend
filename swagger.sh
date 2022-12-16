for d in server/proto/* ; do
    [ -L "${d%/}" ] && continue
    protoc -I=. -I/usr/local/include -I="$GOPATH"/src/github.com/grpc-ecosystem/grpc-gateway -I="$GOPATH"/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. --swagger_out=logtostderr=true:./swagger "$d"/*.proto
    echo "$d - BUILD SUCCESS!"
done

for d in swagger/server/proto/* ; do
    [ -L "${d%/}" ] && continue
    cp -r "$d"/* swagger
done

rm -r swagger/server