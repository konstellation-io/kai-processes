package main

import (
	"context"
	"net"
	"os/signal"
	"syscall"
	"time"

	context2 "context"

	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"google.golang.org/grpc"

	//triggerpb "github.com/konstellation-io/kai-processes/grpc-trigger/proto"
	triggerpb "grpc_trigger/proto"
)

func main() {
	runner.
		NewRunner().
		TriggerRunner().
		WithInitializer(func(ctx sdk.KaiSDK) {
			ctx.Logger.Info("Initializer")
		}).
		WithRunner(grpcServerRunner).
		WithFinalizer(func(ctx sdk.KaiSDK) {
			ctx.Logger.Info("Finalizer")
		}).
		Run()
}

type server struct {
	triggerpb.UnimplementedGreeterServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) SayHello(ctx context.Context, in *triggerpb.Request) (*triggerpb.HelloReply, error) {
	return &triggerpb.HelloReply{Message: in.Name + " world"}, nil
}

func grpcServerRunner(tr *trigger.Runner, sdk sdk.KaiSDK) {
	sdk.Logger.Info("Starting grpc server", "port", 8080)

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		sdk.Logger.Error(err, "Failed to listen")
	}

	bgCtx, stop := signal.NotifyContext(context2.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a gRPC server object
	srv := grpc.NewServer()

	// Serve gRPC Server
	sdk.Logger.Info("Serving gRPC on 0.0.0.0:8080")
	go func() {
		if err := srv.Serve(lis); err != nil {
			sdk.Logger.Error(err, "Failed to serve")
		}
	}()

	<-bgCtx.Done()
	stop()
	sdk.Logger.Info("Shutting down server...")

	// The sdk is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	bgCtx, cancel := context2.WithTimeout(context2.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(bgCtx); err != nil {
		sdk.Logger.Error(err, "Error shutting down server")
	}

	sdk.Logger.Info("Server stopped")
}
