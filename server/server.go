package server

import (
	"math"
	"net"
	"time"

	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/gateway"
	admin_pbaccount "github.com/cryptogateway/backend-envoys/server/proto/v1/admin.pbaccount"
	admin_pbads "github.com/cryptogateway/backend-envoys/server/proto/v1/admin.pbads"
	admin_pbmarket "github.com/cryptogateway/backend-envoys/server/proto/v1/admin.pbmarket"
	admin_pbspot "github.com/cryptogateway/backend-envoys/server/proto/v1/admin.pbspot"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbaccount"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbads"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbauth"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbfuture"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbindex"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbkyc"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbprovider"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbspot"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbstock"
	admin_account "github.com/cryptogateway/backend-envoys/server/service/v1/admin.account"
	admin_ads "github.com/cryptogateway/backend-envoys/server/service/v1/admin.ads"
	admin_market "github.com/cryptogateway/backend-envoys/server/service/v1/admin.market"
	admin_spot "github.com/cryptogateway/backend-envoys/server/service/v1/admin.spot"
	"github.com/cryptogateway/backend-envoys/server/service/v2/account"
	"github.com/cryptogateway/backend-envoys/server/service/v2/ads"
	"github.com/cryptogateway/backend-envoys/server/service/v2/auth"
	"github.com/cryptogateway/backend-envoys/server/service/v2/future"
	"github.com/cryptogateway/backend-envoys/server/service/v2/index"
	"github.com/cryptogateway/backend-envoys/server/service/v2/kyc"
	"github.com/cryptogateway/backend-envoys/server/service/v2/provider"
	"github.com/cryptogateway/backend-envoys/server/service/v2/spot"
	"github.com/cryptogateway/backend-envoys/server/service/v2/stock"
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
)

// Register - The purpose of this function is to create a gRPC server with certain options and to define a gateway for it. It sets
// up a channel to listen on, creates TLS credentials, adds an interceptor for all, creates an array of gRPC options with
// the credentials, registers the handler object, runs a spot service, registers reflection, serves and listens, and sets
// up a gateway. It also uses Logrus entry and grpcctxtags for pre-definition of certain fields by the user.
func Register(option *assets.Context) {

	// MuxOptions is a variable that is used to store a collection of runtime.ServeMuxOption values. Runtime.ServeMuxOption
	// is an interface that allows customization of the ServeMux when creating a new server. It is used to configure various
	// aspects of the ServeMux, such as default routes, route matchers, and middleware.
	var (
		MuxOptions []runtime.ServeMuxOption
	)

	// The option.Write() method is used to write the contents of an Option object to a file. It is used to save changes
	// made to an Option object to a file, so that the same changes can be accessed later.
	option.Write()

	// This is an anonymous function that is being called. The purpose of this function is to execute code asynchronously
	// with the main program. It takes in a pointer to an assets context as an argument, which can then be accessed by the
	// code inside the function. This allows the code inside the function to access and modify data from the main program.
	go func(option *assets.Context) {

		// The purpose of grpclogrus.ReplaceGrpcLogger is to replace the default gRPC logger with a custom logger. In this
		// case, it replaces the gRPC logger with a new logrus entry created from the given option.Logger. This allows the user
		// to use a custom logger, such as logrus, for logging gRPC messages.
		grpclogrus.ReplaceGrpcLogger(logrus.NewEntry(option.Logger))

		// This code is used to establish a connection between the server and the client. The net.Listen() function takes in
		// two parameters, a transport protocol (in this case "tcp") and a host address. It returns a Listener interface and an
		// error. If the error is not nil, the option.Logger.Fatal() function is used to log the error and terminate the program.
		lis, err := net.Listen("tcp", option.Server.Host)
		if err != nil {
			option.Logger.Fatal(err)
		}

		// This code is used to create a new server TLS from file and assign it to the variable "certificate". If an error
		// occurs, the option.Logger.Fatal(err) is called to log the error and the program execution is terminated.
		certificate, err := credentials.NewServerTLSFromFile(option.Credentials.Crt, option.Credentials.Key)
		if err != nil {
			option.Logger.Fatal(err)
		}

		// This code is creating an array of gRPC server options which will be used to configure a gRPC server. The options
		// include setting up a certificate, keepalive parameters, keepalive enforcement policy, a connection timeout, unary
		// and stream server chains with various configurations, and the maximum concurrent streams. All of these options are
		// necessary for the gRPC server to be configured properly and securely.
		opts := []grpc.ServerOption{

			// grpc.Creds(certificate) is a function that is used to create a secure TLS (Transport Layer Security) credential
			// object that can be used to authenticate secure connections to a server. The certificate parameter is used to
			// provide the server's public certificate which is used to verify the server's identity.
			grpc.Creds(certificate),

			// The grpc.KeepaliveParams function is used to configure the keepalive parameters for a gRPC connection. In this
			// instance, it is setting the maximum idle time of a connection to 5 minutes, sending pings every 10 seconds when
			// there is no activity, and waiting 1 second for a ping ack before considering the connection dead. This helps to
			// keep the connection alive, even when there is no activity on it, and ensure that any network issues that arise are quickly detected.
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle: 5 * time.Minute,        // The maximum idle time of this connection will be released if it exceeds, and the proxy will wait until the network problem is solved (the grpc client and server are not notified).
				Time:              10 * time.Second,       // Send pings every 10 seconds if there is no activity.
				Timeout:           100 * time.Millisecond, // Wait 1 second for ping ack before considering the connection dead.
			}),

			// This grpc.KeepaliveEnforcementPolicy sets the parameters for enforcing a keepalive policy for a gRPC connection. It
			// ensures that the client will send a ping at least every 10 seconds or the connection will be terminated. It also
			// allows for pings even when there are no active streams, which is necessary for ensuring that the connection is
			// still active and healthy.
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             10 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection.
				PermitWithoutStream: true,             // Allow pings even when there are no active streams.
			}),

			// The grpc.ConnectionTimeout() function is used to set the maximum amount of time for the connection to be
			// established when establishing a connection to a gRPC server. It takes in a single argument of type time.Duration
			// which specifies the amount of time in seconds to wait for the connection to be established before timing out.
			grpc.ConnectionTimeout(time.Second),

			// The grpcmiddleware.WithUnaryServerChain is used to chain multiple server-side interceptors together. This allows for
			// multiple functions to be executed for a single request, such as authentication, authorization, logging, and other
			// functions. By chaining interceptors together, developers can create powerful and efficient server-side applications.
			grpcmiddleware.WithUnaryServerChain(

				// The grpcctxtags.UnaryServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)) is a
				// function used to extract fields from a gRPC request. It allows the request to be inspected and provides context to
				// the server. This is used to add custom fields to the gRPC context, which can be used for logging and other purposes.
				grpcctxtags.UnaryServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),

				// This UnaryServerInterceptor is a grpc middleware that allows for logging with Logrus. It takes a Logger option as
				// well as a list of grpclogrus options to create a Logrus entry. The purpose of this is to intercept requests, log
				// them using the Logrus logger, and then pass the request on to the server.
				grpclogrus.UnaryServerInterceptor(logrus.NewEntry(option.Logger), []grpclogrus.Option{
					grpclogrus.WithLevels(grpclogrus.DefaultCodeToLevel),
				}...),

				// The grpcopentracing.UnaryServerInterceptor() is a function that adds tracing capabilities to a gRPC server. It allows
				// developers to trace requests through the server-side of the application and view performance metrics. This
				// function is particularly useful for debugging, troubleshooting, and optimizing applications.
				grpcopentracing.UnaryServerInterceptor(),

				// The grpc_recovery.UnaryServerInterceptor is a gRPC interceptor that adds support for recovery from panics in gRPC
				// server applications. It provides an option to specify a custom recovery handler, which is what the
				// WithRecoveryHandler option does. This allows developers to customize how the recovery behaves, such as logging the
				// panic and returning a custom error code.
				grpc_recovery.UnaryServerInterceptor([]grpc_recovery.Option{
					grpc_recovery.WithRecoveryHandler(option.Recovery),
				}...),
			),

			// The grpcmiddleware.WithStreamServerChain(...) is used to create a server-side middleware chain that can be used to intercept and modify requests and responses on a gRPC server.
			// It allows developers to add functionality such as authentication, logging, and monitoring to the server.
			grpcmiddleware.WithStreamServerChain(

				// This is a gRPC StreamServerInterceptor which is used to extract fields from the incoming request and add them as
				// context tags. The CodeGenRequestFieldExtractor is used to extract the fields from the request. This allows the
				// user to quickly access the request fields without having to manually parse them.
				grpcctxtags.StreamServerInterceptor(grpcctxtags.WithFieldExtractor(grpcctxtags.CodeGenRequestFieldExtractor)),

				// The grpclogrus.StreamServerInterceptor is a function used to add logging to a gRPC server. It takes a
				// logrus.NewEntry, which is a log entry with predefined fields, and a list of grpclogrus.Option parameters which are
				// used to configure the logging. The main purpose of this function is to provide extra logging information for debugging purposes.
				grpclogrus.StreamServerInterceptor(logrus.NewEntry(option.Logger), []grpclogrus.Option{
					grpclogrus.WithLevels(grpclogrus.DefaultCodeToLevel),
				}...),

				// grpcopentracing.StreamServerInterceptor() is a function that is used to create a stream server interceptor for
				// OpenTracing. The purpose of the interceptor is to provide distributed tracing capabilities for the gRPC server.
				// The interceptor is responsible for automatically tracing incoming requests and generating trace spans for each
				// request. This allows developers to better monitor and debug their gRPC applications.
				grpcopentracing.StreamServerInterceptor(),

				// The grpc_recovery.StreamServerInterceptor is a middleware that allows for recovery from unexpected panics that
				// occur during the handling of gRPC requests. The grpc_recovery.WithRecoveryHandler() option specifies a function
				// that will be called when a panic is encountered, allowing the server to recover and continue processing requests.
				// This can help ensure that the server remains stable in the face of unexpected errors.
				grpc_recovery.StreamServerInterceptor([]grpc_recovery.Option{
					grpc_recovery.WithRecoveryHandler(option.Recovery),
				}...),
			),

			// The purpose of grpc.MaxConcurrentStreams(math.MaxUint32) is to set the maximum number of concurrent streams to the
			// maximum value supported by the protocol, which is 2^32 - 1. This allows for an unlimited number of concurrent
			// streams, as long as the server can handle them. This is useful for applications that need to handle a high volume of requests.
			grpc.MaxConcurrentStreams(math.MaxUint32),
		}

		// The purpose is to create a new grpc server with the given options. This server can
		// then be used to handle incoming requests from clients.
		srv := grpc.NewServer(opts...)

		serviceSpot := spot.Service{Context: option}
		serviceSpot.Initialization()
		pbspot.RegisterApiServer(srv, &serviceSpot)

		serviceProvider := provider.Service{Context: option}
		serviceProvider.Initialization()
		pbprovider.RegisterApiServer(srv, &provider.Service{Context: option})

		pbstock.RegisterApiServer(srv, &stock.Service{Context: option})
		pbindex.RegisterApiServer(srv, &index.Service{Context: option})
		pbauth.RegisterApiServer(srv, &auth.Service{Context: option})
		pbaccount.RegisterApiServer(srv, &account.Service{Context: option})
		pbads.RegisterApiServer(srv, &ads.Service{Context: option})
		pbkyc.RegisterApiServer(srv, &kyc.Service{Context: option})
		// serviceFuture := future.Service{Context: option}
		pbfuture.RegisterApiServer(srv, &future.Service{Context: option})

		admin_pbaccount.RegisterApiServer(srv, &admin_account.Service{Context: option})
		admin_pbads.RegisterApiServer(srv, &admin_ads.Service{Context: option})
		admin_pbspot.RegisterApiServer(srv, &admin_spot.Service{Context: option})
		admin_pbmarket.RegisterApiServer(srv, &admin_market.Service{Context: option})

		// Reflection.Register is a method that registers a service with a gRPC server. This method is used to create a service
		// endpoint to allow clients to communicate with the server. The method sets up a connection between the server and the
		// client, allowing the client to make calls to the server. The service can then respond to the client's requests.
		reflection.Register(srv)

		// This code is used to start a server and listen for incoming connections on a specified network address. The Serve()
		// method is used to start the server and it takes a listener as an argument. If an error is encountered while starting
		// the server, the error is logged using the Logger object.
		if err := srv.Serve(lis); err != nil {
			option.Logger.Fatal(err)
		}

	}(option)

	// This code is used to append a JSONPb marshaler to the MuxOptions slice if option.Development is true. The JSONPb
	// marshaler is used to encode and decode messages in the JSON format. This marshaler also allows for the original
	// naming of fields to be preserved, as well as providing a more readable "indented" format for the output JSON.
	if option.Development {
		MuxOptions = append(MuxOptions, runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName: true,
			Indent:   "   ",
		}))
	}

	// This code is used to create an HTTP/2 gateway server which allows clients to connect and access a gRPC service. The
	// gateway.Run() function sets up the server and its options, including the address (option.Server.Proxy), the gRPC
	// server address (option.Server.Host), the context (option), and the mux (MuxOptions). If an error occurs during the
	// setup, it is logged with option.Logger.Fatal().
	if err := gateway.Run(gateway.Options{

		// The Addr option.Server.Proxy is used to set the address of the proxy server that will be used when making requests.
		// This allows the server to route requests to different servers depending on the configuration of the proxy.
		Addr: option.Server.Proxy,

		// GRPCServer is an endpoint defined in the gateway that defines the address of the server. The purpose of this
		// endpoint is to provide a way for clients to connect to the server, allowing for requests and responses to be sent and received.
		GRPCServer: gateway.Endpoint{
			Addr: option.Server.Host,
		},
		Context: option,
		Mux:     MuxOptions,
	}); err != nil {
		option.Logger.Fatal(err)
	}

}
