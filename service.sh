name=$(echo "$1" | tr '[:upper:]' '[:lower:]')

if [ -d "server/proto/pb$name" ]; then
  echo "Service $name already exists. Stopping the process." >&2
  exit 1
fi

mkdir "server/proto/pb$name"
touch "server/proto/pb$name/$name.proto"
echo 'syntax = "proto3";' >> "server/proto/pb$name/$name.proto"
echo "package pb$name;" >> "server/proto/pb$name/$name.proto"
echo 'option go_package = "server/proto/pb'"$name"'";' >> server/proto/pb"$name"/"$name".proto
echo 'import "google/api/annotations.proto";' >> "server/proto/pb$name/$name.proto"

# shellcheck disable=SC2129
echo 'service Api {' >> server/proto/pb"$name"/"$name".proto
echo '  rpc GetTest (GetRequestTest) returns (ResponseTest) {' >> server/proto/pb"$name"/"$name".proto
echo '    option (google.api.http) = {' >> server/proto/pb"$name"/"$name".proto
echo '      post: "/v2/'"$name"'/get-test",' >> server/proto/pb"$name"/"$name".proto
echo '      body: "*"' >> server/proto/pb"$name"/"$name".proto
echo '    };' >> server/proto/pb"$name"/"$name".proto
echo '  }' >> server/proto/pb"$name"/"$name".proto
echo '}' >> server/proto/pb"$name"/"$name".proto

echo 'message GetRequestTest {}' >> server/proto/pb"$name"/"$name".proto
echo 'message ResponseTest {}' >> server/proto/pb"$name"/"$name".proto

echo "+++++++++ Create Proto file +++++++++"
echo "server/proto/pb$name/$name.proto"
echo "++++++++++ Create Service +++++++++++"

mkdir "server/service/$name"
touch "server/service/$name/$name.go"
echo "package $name" >> "server/service/$name/$name.go"
touch "server/service/$name/$name.grpc.go"
echo "package $name" >> "server/service/$name/$name.grpc.go"
touch "server/service/$name/$name.admin.grpc.go"
echo "package $name" >> "server/service/$name/$name.admin.grpc.go"

# shellcheck disable=SC2129
echo 'import "github.com/cryptogateway/backend-envoys/assets"' >> "server/service/$name/$name.go"
echo "type Service struct {" >> "server/service/$name/$name.go"
echo "	Context *assets.Context" >> "server/service/$name/$name.go"
echo "}" >> "server/service/$name/$name.go"

# shellcheck disable=SC2129
echo 'import (' >> "server/service/$name/$name.grpc.go"
echo '  "context"' >> "server/service/$name/$name.grpc.go"
echo '  "github.com/cryptogateway/backend-envoys/server/proto/pb'"$name"'"' >> "server/service/$name/$name.grpc.go"
echo ')' >> "server/service/$name/$name.grpc.go"

echo 'func (s Service) GetTest(ctx context.Context, test *pb'"$name"'.GetRequestTest) (*pb'"$name"'.ResponseTest, error) {' >> "server/service/$name/$name.grpc.go"
echo '  panic("implement me")' >> "server/service/$name/$name.grpc.go"
echo '}' >> "server/service/$name/$name.grpc.go"

echo "server/service/$name/$name.go"
echo "server/service/$name/$name.grpc.go"
echo "server/service/$name/$name.admin.grpc.go"

echo "+++++++++ Register Service ++++++++++"

search=$(grep -n "srv := grpc.NewServer(opts...)" ./server/server.go)
line1=$(echo "$search" | cut -d':' -f1)
sed -i "${line1}a\\\t\tpb$name.RegisterApiServer(srv, &$name.Service{Context: option})" ./server/server.go

echo "Done!."

echo "+++++++++ Register gateway ++++++++++"

search=$(grep -n "for _, f := range \[\]func(context.Context, \*runtime.ServeMux, \*grpc.ClientConn) error{" ./server/gateway/gateway.go)
line2=$(echo "$search" | cut -d':' -f1)
sed -i "${line2}a\\\t\tpb$name.RegisterApiHandler," ./server/gateway/gateway.go

echo "Done!."

echo "+++++++++++ Compile proto +++++++++++"

bash ./proto.sh