package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cryptogateway/backend-envoys/assets"
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
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Endpoint describes a gRPC endpoint.
type Endpoint struct {
	Addr string
}

// Options - The purpose of this struct is to provide a set of options that can be used to configure a gRPC service. The Addr field
// provides the address to listen on, the GRPCServer field provides the endpoint of the gRPC service, the Mux field
// provides a list of options to be passed to the grpc-gateway multiplexer, the Context field provides a full assets
// context, and the Certificate field provides a TLS certificate.
type Options struct {

	// Addr is the address to listen.
	Addr string

	// GRPCServer Endpoint is a point of entry into a GRPC Server that enables remote procedure calls to be made between
	// client applications and the server. It provides an interface for clients to call functions on the server, receive
	// results, and send parameters. The endpoint is a key part of the GRPC architecture and helps facilitate communication
	// between the client and server.
	GRPCServer Endpoint

	// Mux are runtime.ServeMuxOptions that allow for customizing the behavior of a ServeMux. They provide a way to
	// configure the ServeMux without needing to modify the source code. For example, a ServeMuxOption can be used to set
	// the timeout for requests or set the logging level.
	Mux []runtime.ServeMuxOption

	// The context *assets.Context is used to store information about the assets that are being used in an application.
	// It is used to track the various assets used in the application and make sure that they are kept up to date.
	// It can also be used to store data related to the assets such as metadata, versions, and references.
	Context *assets.Context

	// The purpose of the tls.Certificate is to provide a secure encryption protocol for data transmissions over a network.
	// It provides a secure way for two parties to communicate with each other, ensuring the data is not intercepted or
	// tampered with. The TLS (Transport Layer Security) certificate provides a way to authenticate the server, as well as
	// encrypting the data sent between the two entities.
	Certificate tls.Certificate
}

// Run - The purpose of this code snippet is to establish a connection between a client and server using the gRPC protocol. It
// creates a context with a cancel function that can be used to stop the goroutine, configures the gRPC connection, sets
// up keep-alive parameters, sets the maximum size of messages that the server can receive, and sets the headers for the
// context. Finally, it closes the gRPC client connection when the context is canceled.
func Run(params Options) error {

	// This code snippet is used to create a context with a cancel function that can be used to stop a goroutine. The
	// context allows for the sharing of request-scoped values and cancellation signals across a group of goroutines. The
	// cancel function that is returned from the WithCancel call is used to stop the goroutine by sending a cancellation
	// signal. The defer statement ensures that the cancel function is always called when the function that the snippet is
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the TLS credentials.
	certificate, err := credentials.NewClientTLSFromFile(params.Context.Credentials.Crt, params.Context.Credentials.Override)
	if err != nil {
		return err
	}

	// The opts variable is a slice of grpc.DialOption values. The purpose of these two options is to configure the gRPC connection.
	// WithInsecure() makes sure that the connection is not secure and WithBlock() ensures that the connection blocks until the underlying connection is up and running.
	opts := []grpc.DialOption{

		// grpc.WithTransportCredentials(certificate) is used to specify the credentials to use when making an RPC call. The
		// certificate provided as an argument is used to authenticate the client making the request, ensuring that the server
		// only allows requests from an authorized source.
		grpc.WithTransportCredentials(certificate),

		// This code is used to set up keep-alive parameters for a gRPC client. Keep-alive is used to maintain a connection
		// between the client and the server, even when there is no activity. Setting this up allows the client to send a ping
		// to the server every 10 seconds if there is no activity, and wait for 100 milliseconds for the server to respond
		// before considering the connection dead. The permit without stream option allows the client to send pings even
		// without active streams.
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,       // Send pings every 10 seconds if there is no activity.
			Timeout:             100 * time.Millisecond, // Wait 100 millisecond for ping ack before considering the connection dead.
			PermitWithoutStream: true,                   // Send pings even without active streams.
		}),
		grpc.WithBlock(),

		// The purpose of grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4194304)) is to set the maximum size of the
		// messages that the gRPC server can receive. This ensures that large messages do not overload the server and cause it
		// to crash. It also helps to prevent malicious users from sending large messages in order to exhaust the server's resources.
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4194304)),
	}

	// This code is establishing a connection between a client and server using the gRPC protocol. The
	// `params.Context.GrpcClient` variable is being assigned a new gRPC client connection, and the `params.GRPCServer.Addr`
	// variable is the address of the gRPC server. The `opts...` variable are optional arguments for the gRPC DialContext
	// method. If the connection cannot be established, an error is returned.
	params.Context.GrpcClient, err = grpc.DialContext(ctx, params.GRPCServer.Addr, opts...)
	if err != nil {
		return err
	}

	// This code snippet is used to close a connection that is established via a GRPC client. The code is wrapped in a go
	// routine which will be executed as a separate goroutine. The context.Done() method is used to block until the context
	// is canceled and then close the GRPC client connection. If an error occurs while closing the client, it will be logged
	// using the context.Logger.Fatal() method.
	go func() {
		<-ctx.Done()
		if err := params.Context.GrpcClient.Close(); err != nil {
			params.Context.Logger.Fatal(err)
		}
	}()

	// This code is used to set the headers for a context and check for any errors that might occur. If an error occurs, the
	// code will return the error.
	if err := params.headers(ctx, params.Context.GrpcClient); err != nil {
		return err
	}

	return nil
}

// The purpose of this function is to write an error message to the http ResponseWriter with a specified HTTP status
// code. It sends an error response indicating that the request was not successful and provides the user with information about the error.
func (o *Options) error(w http.ResponseWriter, _ error, status int) {
	w.WriteHeader(status)
}

func (o *Options) headers(ctx context.Context, conn *grpc.ClientConn) error {

	// The purpose of this code is to create a new multiplexer for HTTP requests. A ServeMux is an HTTP request multiplexer
	// which matches the URL of incoming requests against a list of registered paths and calls a corresponding handler for
	// the path. The http.NewServeMux() function creates and returns a new ServeMux object which can be used to route
	// requests to different handlers.
	route := http.NewServeMux()

	// The purpose of this code is to create a route that will serve static files from the ./static directory when a request
	// is made to the /v2/storage/ URL. This route is created by using the http.StripPrefix function which will strip the
	// "/v2/storage/" portion from the URL, allowing http.FileServer to serve the files from the ./static directory.
	route.Handle("/v2/storage/", http.StripPrefix("/v2/storage/", http.FileServer(http.Dir("./static"))))

	// The purpose of this code is to send an HTTP response to the client with a plain text message, indicating the state of the grpc server.
	// If the state of the server is not "Ready", an error message is sent to the client indicating the state of the server. If the state is "Ready", an "ok" message is sent to the client.
	route.HandleFunc("/v2/status", func(conn *grpc.ClientConn) http.HandlerFunc {

		// This code is a function that is used to check the state of a grpc server. It is written in the Go programming
		// language. The purpose of this code is to send a http response to the client with a plain text message, indicating
		// the state of the grpc server. If the state of the server is not "Ready", an http error message is sent to the client
		// indicating the state of the server. If the state is "Ready", an "ok" message is sent to the client.
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			if s := conn.GetState(); s != connectivity.Ready {
				http.Error(w, fmt.Sprintf("grpc server is %s", s), http.StatusBadGateway)
				return
			}
			_, _ = fmt.Fprintln(w, "ok")
		}

	}(conn))

	// The route.HandleFunc() function is used to register a handler function for a given URL path. In this case, the
	// handler function is used to handle requests to the "/v2/timestamp" URL path. This handler function takes a
	// grpc.ClientConn as its argument and returns a http.HandlerFunc. The http.HandlerFunc is responsible for handling
	// HTTP requests and responses for the given URL path.
	route.HandleFunc("/v2/timestamp", func(conn *grpc.ClientConn) http.HandlerFunc {

		// This code is a function that is used to write the current Unix timestamp to an HTTP response. The function takes an
		// HTTP response writer and an HTTP request as parameters, and it uses the fmt package to write the current Unix
		// timestamp to the response.
		return func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, time.Now().UTC().Unix())
		}

	}(conn))

	// This code is attempting to set up a gateway connection and will return an error if it fails. The gateway variable is
	// being assigned the result of the o.gateway function which will be used to establish a connection. The context,
	// connection and multiplexer are being passed to the function as parameters. If the gateway connection fails, the code
	// will return an error.
	gateway, err := o.gateway(ctx, conn, o.Mux)
	if err != nil {
		return err
	}

	// The purpose of this code is to register a handler for a given route. In this case, the route is "/" and the handler
	// is "gateway". This means that when a request is made to route "/", the "gateway" handler will be executed.
	route.Handle("/", gateway)

	// The purpose of this code is to configure an HTTP server using Go's net/http package.
	// The server will listen on port 8080 and use the provided handler (myHandler) to handle incoming requests.
	// The server will also have a read and write timeout of 10 seconds and a maximum header size of 1MB.
	s := &http.Server{
		Addr: o.Addr,

		// This is a function that takes an http.Handler as an argument and returns another http.Handler. It is typically used
		// for adding additional functionality to an existing handler, such as logging, authentication, or rate limiting.
		Handler: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// The above code is a part of a function that is used to set the Access-Control-Allow-Origin header. It checks to
				// see if the Origin header was sent in the request and if so, it stores it in the variable origin. The value of
				// origin will then be used in the response to set the Access-Control-Allow-Origin header. This allows the browser to
				// know which domains are allowed to access the resource.
				if origin := r.Header.Get("Origin"); origin != "" {

					// This line of code is used to set an HTTP header for Cross-Origin Resource Sharing (CORS). It adds the
					// "Access-Control-Allow-Origin" header to the response, with the specified origin as the value. This allows the
					// browser to determine if a cross-origin request should be allowed.
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// The purpose of the code above is to set the headers in the response of a web request. It is used to specify the
					// list of allowed headers that the browser will accept when making a request to the server. This code will allow
					// browsers to send requests with the specified headers to the server, which makes the process of requesting data
					// more secure.
					w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "X-Accept-Content-Transfer-Encoding", "X-Accept-Response-Streaming", "X-User-Agent", "X-Grpc-Web", "Message-Encoding", "Message-Accept-Encoding", "Message-Type", "Timeout"}, ","))

					// The purpose of the code is to set the Access-Control-Allow-Methods header in an HTTP response to include the
					// methods "GET", "OPTIONS", and "POST". This allows the server to specify which methods are allowed when making
					// requests to the server. This is often used in Cross-Origin Resource Sharing (CORS) requests.
					w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{"GET", "OPTIONS", "POST"}, ","))

					// The purpose of this code is to set the header of an HTTP response to "Content-Type" with a value of
					// "Application/grpc". This header is used to indicate the type of content that is being sent in the response body.
					w.Header().Set("Content-Type", "application/grpc")

					// The purpose of this statement is to set the Access-Control-Max-Age header. This header is used for CORS
					// (Cross-Origin Resource Sharing) and limits the amount of time that a browser can cache the preflight response.
					// This helps to prevent malicious attacks that can be performed by caching the preflight response.
					w.Header().Set("Access-Control-Max-Age", "3600")

					// This code is used to handle an OPTIONS request which is sent by the browser when a cross-origin request is made.
					// The browser will send an OPTIONS request with the Access-Control-Request-Method header to determine if the server
					// will allow the cross-origin request. If the request is an OPTIONS request with the Access-Control-Request-Method
					// header, then the code returns without doing anything else.
					if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
						return
					}

					// This code is used for Cross-Origin Resource Sharing (CORS). The purpose of this code is to specify the headers
					// that the server should expose to the client. When the request method is a POST and the request header contains an
					// Access-Control-Request-Method, this code sets the Access-Control-Expose-Headers header to a list of strings,
					// including Content-Transfer-Encoding, Grpc-Message, and Grpc-Status. This allows the server to expose the
					// specified headers to the client.
					if r.Method == "POST" && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Expose-Headers", strings.Join([]string{"Content-Transfer-Encoding", "Grpc-Message", "Grpc-Status"}, ","))
						return
					}
				}

				//ServeHTTP is a method of the http.Handler interface. It is responsible for writing the response for a given HTTP
				//request. It takes two parameters, a http.ResponseWriter and a http.Request, and uses them to write the response.
				h.ServeHTTP(w, r)
			})
		}(route),

		// TLSConfig is a function that is used to set up a secure connection between two parties by creating a TLS (Transport
		// Layer Security) configuration. It takes an Options type as an argument and returns a tls.Config type. The tls.Config
		// type contains the parameters necessary for the secure connection, such as the protocol version, the certificate, and the cipher suite.
		TLSConfig: func(o *Options) *tls.Config {

			// This code is reading a file containing a certificate, and storing it in the variable cert. It is also checking for
			// any errors that may occur during the read operation and logging them. This code is likely part of an authentication
			// process or some kind of secure communication.
			cert, err := ioutil.ReadFile(o.Context.Credentials.Crt)
			if err != nil {
				o.Context.Logger.Error(err)
			}

			// This code is attempting to read a file containing credentials from the o.Context.Credentials.Key location and
			// assign it to the key variable. The if statement is checking for any errors that may occur during the process, and
			// if so, the error is logged to the o.Context.Logger.
			key, err := ioutil.ReadFile(o.Context.Credentials.Key)
			if err != nil {
				o.Context.Logger.Error(err)
			}

			// This code is used to create an X509 certificate from two files: cert and key. It is then assigned to the
			// Certificate variable. If an error occurs, it is logged to the Context.Logger.
			o.Certificate, err = tls.X509KeyPair(cert, key)
			if err != nil {
				o.Context.Logger.Error(err)
			}

			// This code is used to configure a TLS connection with the server. It sets up the TLS configuration with the
			// appropriate certificate and protocol for the connection. The tls.Config struct is then returned which can be used
			// to set up the connection with the server.
			return &tls.Config{
				Certificates: []tls.Certificate{o.Certificate},
				NextProtos:   []string{http2.NextProtoTLS},
			}
		}(o),
	}

	// This function is used to gracefully shut down an HTTP server given a set of options, an HTTP server, and a context.
	// The function listens for the context to be done, then logs that the server is being shut down. It then attempts to
	// shut down the server using the context provided. If the shutdown fails, an error is logged.
	go func(o *Options, s *http.Server, ctx context.Context) {
		<-ctx.Done()
		o.Context.Logger.Infof("Shutting down the http server")

		// This code is used to shut down a http server. The "if err" statement is used to check for any errors that may occur
		// during the shutdown process. If an error is encountered, the code will log the error using the
		// Context.Logger.Errorf() function.
		if err := s.Shutdown(context.Background()); err != nil {
			o.Context.Logger.Errorf("Failed to shutdown http server: %v", err)
		}
	}(o, s, ctx)

	// The purpose of this code is to start a web server listening on the address specified in the "o.Addr" variable. The
	// first line is a logging statement to indicate that the server is starting to listen at the specified address. The
	// second line attempts to start the server and logs an error if it fails to start.
	o.Context.Logger.Infof("Starting listening at %s", o.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		o.Context.Logger.Errorf("Failed to listen and serve: %v", err)
		return err
	}

	return nil
}

// This function is used to define and register a gateway between a gRPC client and server, allowing for communication
// between the two. It takes a context, a gRPC Client connection and a list of ServeMux options as arguments and returns
// an HTTP Handler and an error value. It then registers various API handlers for communication between the client and server.
func (o *Options) gateway(ctx context.Context, connect *grpc.ClientConn, opts []runtime.ServeMuxOption) (http.Handler, error) {

	// The purpose of this code is to create a ServeMux object using the runtime package. The ServeMux is an HTTP request
	// multiplexer that matches the URL of each incoming request against a list of registered patterns and calls the handler
	// for the pattern that most closely matches the URL. The opts parameter is an optional variadic list of options that
	// are used to configure the ServeMux.
	route := runtime.NewServeMux(opts...)

	// This code is a loop that is registering a list of API handlers to a given context, ServeMux, and ClientConn. It is
	// used to associate the API handlers with the given context, ServeMux, and ClientConn so that they can be used to
	// handle requests within the application. After registering each API handler, the loop checks for any errors that may
	// have occurred during the registration process. If an error occurs, the loop will return an error and the registration
	// process will be terminated.
	for _, f := range []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{
		// V2 - Public apis
		pbindex.RegisterApiHandler,
		pbauth.RegisterApiHandler,
		pbaccount.RegisterApiHandler,
		pbspot.RegisterApiHandler,
		pbads.RegisterApiHandler,
		pbstock.RegisterApiHandler,
		pbkyc.RegisterApiHandler,
		pbprovider.RegisterApiHandler,
		pbfuture.RegisterApiHandler,
		// V1 - Admin apis.
		admin_pbaccount.RegisterApiHandler,
		admin_pbspot.RegisterApiHandler,
		admin_pbads.RegisterApiHandler,
		admin_pbmarket.RegisterApiHandler,
	} {
		if err := f(ctx, route, connect); err != nil {
			return nil, err
		}
	}

	return route, nil
}
