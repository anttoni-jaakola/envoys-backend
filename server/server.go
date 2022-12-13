package server

import (
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/gateway"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
	"github.com/cryptogateway/backend-envoys/server/proto/pbauth"
	"github.com/cryptogateway/backend-envoys/server/proto/pbindex"
	"github.com/cryptogateway/backend-envoys/server/proto/pbmarket"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/service/account"
	"github.com/cryptogateway/backend-envoys/server/service/auth"
	"github.com/cryptogateway/backend-envoys/server/service/index"
	"github.com/cryptogateway/backend-envoys/server/service/market"
	"github.com/cryptogateway/backend-envoys/server/service/spot"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpclogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcopentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"math"
	"net"
	"time"
)

func Master(option *assets.Context) {

	var (
		MuxOptions []runtime.ServeMuxOption
	)

	option.Write()

	// Independent layer.
	go func(option *assets.Context) {

		// Logrus entry is used, allowing pre-definition of certain fields by the user.
		// Make sure that log statements internal to gRPC library are logged using the logrus Logger as well.
		grpclogrus.ReplaceGrpcLogger(logrus.NewEntry(option.Logger))

		// Create the channel to listen on.
		lis, err := net.Listen("tcp", option.GrpcAddress)
		if err != nil {
			option.Logger.Fatal(err)
		}

		// Create the TLS credentials.
		certificate, err := credentials.NewServerTLSFromFile(option.CredentialsCrt, option.CredentialsKey)
		if err != nil {
			option.Logger.Fatal(err)
		}

		// The following grpc.ServerOption adds an interceptor for all.
		// Create an array of gRPC options with the credentials.
		opts := []grpc.ServerOption{
			grpc.Creds(certificate),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle: 5 * time.Minute,        // The maximum idle time of this connection will be released if it exceeds, and the proxy will wait until the network problem is solved (the grpc client and server are not notified).
				Time:              10 * time.Second,       // Send pings every 10 seconds if there is no activity.
				Timeout:           100 * time.Millisecond, // Wait 1 second for ping ack before considering the connection dead.
			}),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             10 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection.
				PermitWithoutStream: true,             // Allow pings even when there are no active streams.
			}),
			grpc.ConnectionTimeout(time.Second),
			grpcmiddleware.WithUnaryServerChain(
				grpcctxtags.UnaryServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),
				grpclogrus.UnaryServerInterceptor(logrus.NewEntry(option.Logger), []grpclogrus.Option{
					grpclogrus.WithLevels(grpclogrus.DefaultCodeToLevel),
				}...),
				grpcopentracing.UnaryServerInterceptor(),
				grpc_recovery.UnaryServerInterceptor([]grpc_recovery.Option{
					grpc_recovery.WithRecoveryHandler(option.Recovery),
				}...),
			),
			grpcmiddleware.WithStreamServerChain(
				grpcctxtags.StreamServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),
				grpclogrus.StreamServerInterceptor(logrus.NewEntry(option.Logger), []grpclogrus.Option{
					grpclogrus.WithLevels(grpclogrus.DefaultCodeToLevel),
				}...),
				grpcopentracing.StreamServerInterceptor(),
				grpc_recovery.StreamServerInterceptor([]grpc_recovery.Option{
					grpc_recovery.WithRecoveryHandler(option.Recovery),
				}...),
			),
			grpc.MaxConcurrentStreams(math.MaxUint32),
		}

		// Create the gRPC server with the credentials.
		srv := grpc.NewServer(opts...)

		// Register the handler object.
		pbindex.RegisterApiServer(srv, &index.Service{Context: option})
		pbauth.RegisterApiServer(srv, &auth.Service{Context: option})
		pbaccount.RegisterApiServer(srv, &account.Service{Context: option})

		var (
			spotService   = &spot.Service{Context: option}
			marketService = &market.Service{Context: option}
		)

		go spotService.Initialization()
		go marketService.Initialization()

		pbmarket.RegisterApiServer(srv, marketService)
		pbspot.RegisterApiServer(srv, spotService)

		reflection.Register(srv)

		// Serve and Listen.
		if err := srv.Serve(lis); err != nil {
			option.Logger.Fatal(err)
		}

	}(option)

	if option.Development {
		MuxOptions = append(MuxOptions, runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName: true,
			Indent:   "   ",
		}))
	}

	if err := gateway.Run(gateway.Options{
		Addr: option.GrpcGatewayAddress,
		GRPCServer: gateway.Endpoint{
			Addr: option.GrpcAddress,
		},
		Context: option,
		Mux:     MuxOptions,
	}); err != nil {
		option.Logger.Fatal(err)
	}

}
