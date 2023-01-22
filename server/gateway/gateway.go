package gateway

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/proto/pbaccount"
	"github.com/cryptogateway/backend-envoys/server/proto/pbauth"
	"github.com/cryptogateway/backend-envoys/server/proto/pbindex"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"github.com/cryptogateway/backend-envoys/server/proto/pbstock"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Endpoint describes a gRPC endpoint.
type Endpoint struct {
	Addr string
}

// Options is a set of options to be passed to Run.
type Options struct {

	// Addr is the address to listen.
	Addr string

	// GRPCServer defines an endpoint of a gRPC service.
	GRPCServer Endpoint

	// Mux is a list of options to be passed to the grpc-gateway multiplexer.
	Mux []runtime.ServeMuxOption

	// Full assets context.
	Context *assets.Context

	// TLS certificate.
	Certificate tls.Certificate
}

// Run starts a HTTP server and blocks while running if successful.
// The server will be shutdown when "ctx" is canceled.
func Run(params Options) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the TLS credentials.
	certificate, err := credentials.NewClientTLSFromFile(params.Context.Credentials.Crt, params.Context.Credentials.Override)
	if err != nil {
		return err
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(certificate),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,       // Send pings every 10 seconds if there is no activity.
			Timeout:             100 * time.Millisecond, // Wait 100 millisecond for ping ack before considering the connection dead.
			PermitWithoutStream: true,                   // Send pings even without active streams.
		}),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4194304)),
	}

	params.Context.GrpcClient, err = grpc.DialContext(ctx, params.GRPCServer.Addr, opts...)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		if err := params.Context.GrpcClient.Close(); err != nil {
			params.Context.Logger.Fatal(err)
		}
	}()

	if err := params.headers(ctx, params.Context.GrpcClient); err != nil {
		return err
	}

	return nil
}

func (o *Options) error(w http.ResponseWriter, err error, status int) {
	//http.Error(w, err.Error(), status); return
	w.WriteHeader(status)
}

func (o *Options) headers(ctx context.Context, conn *grpc.ClientConn) error {

	route := http.NewServeMux()
	route.Handle("/v2/storage/", http.StripPrefix("/v2/storage/", http.FileServer(http.Dir("./static"))))

	// Server status returns a health handler which returns ok.
	route.HandleFunc("/v2/status", func(conn *grpc.ClientConn) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			if s := conn.GetState(); s != connectivity.Ready {
				http.Error(w, fmt.Sprintf("grpc server is %s", s), http.StatusBadGateway)
				return
			}
			_, _ = fmt.Fprintln(w, "ok")
		}
	}(conn))

	route.HandleFunc("/v2/timestamp", func(conn *grpc.ClientConn) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprintln(w, time.Now().UTC().Unix())
		}
	}(conn))

	gateway, err := o.gateway(ctx, conn, o.Mux)
	if err != nil {
		return err
	}

	route.Handle("/", gateway)

	s := &http.Server{
		Addr: o.Addr,

		// Allow CORS allows Cross Origin Resoruce Sharing from any origin.
		// Don't do this without consideration in production systems.
		Handler: func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if origin := r.Header.Get("Origin"); origin != "" {

					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "X-Accept-Content-Transfer-Encoding", "X-Accept-Response-Streaming", "X-User-Agent", "X-Grpc-Web", "Message-Encoding", "Message-Accept-Encoding", "Message-Type", "Timeout"}, ","))
					w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{"GET", "OPTIONS", "POST"}, ","))
					w.Header().Set("Content-Type", "application/grpc")
					w.Header().Set("Access-Control-Max-Age", "3600")

					if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
						return
					}
					if r.Method == "POST" && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Expose-Headers", strings.Join([]string{"Content-Transfer-Encoding", "Grpc-Message", "Grpc-Status"}, ","))
						return
					}
				}
				h.ServeHTTP(w, r)

			})
		}(route),
		// TLS configuration - gRPC has SSL/TLS integration and promotes the use of SSL/TLS to authenticate the server,
		// and to encrypt all the data exchanged between the client and the server.
		// Optional mechanisms are available for clients to provide certificates for mutual authentication.
		TLSConfig: func(o *Options) *tls.Config {
			cert, err := ioutil.ReadFile(o.Context.Credentials.Crt)
			if err != nil {
				o.Context.Logger.Error(err)
			}

			key, err := ioutil.ReadFile(o.Context.Credentials.Key)
			if err != nil {
				o.Context.Logger.Error(err)
			}

			o.Certificate, err = tls.X509KeyPair(cert, key)
			if err != nil {
				o.Context.Logger.Error(err)
			}

			return &tls.Config{
				Certificates: []tls.Certificate{o.Certificate},
				NextProtos:   []string{http2.NextProtoTLS},
			}
		}(o),
	}

	go func(o *Options, s *http.Server, ctx context.Context) {
		<-ctx.Done()
		o.Context.Logger.Infof("Shutting down the http server")
		if err := s.Shutdown(context.Background()); err != nil {
			o.Context.Logger.Errorf("Failed to shutdown http server: %v", err)
		}
	}(o, s, ctx)

	o.Context.Logger.Infof("Starting listening at %s", o.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		o.Context.Logger.Errorf("Failed to listen and serve: %v", err)
		return err
	}

	return nil
}

// The gRPC-Gateway is a plugin of the Google protocol buffers compiler protoc.
// It reads protobuf service definitions and generates a reverse-proxy server which translates a RESTful HTTP API into gRPC.
// This server is generated according to the google.api.http annotations in your service definitions.
func (o *Options) gateway(ctx context.Context, connect *grpc.ClientConn, opts []runtime.ServeMuxOption) (http.Handler, error) {

	route := runtime.NewServeMux(opts...)

	for _, f := range []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{
		pbindex.RegisterApiHandler,
		pbauth.RegisterApiHandler,
		pbaccount.RegisterApiHandler,
		pbspot.RegisterApiHandler,
		pbstock.RegisterApiHandler,
	} {
		if err := f(ctx, route, connect); err != nil {
			return nil, err
		}
	}

	return route, nil
}
