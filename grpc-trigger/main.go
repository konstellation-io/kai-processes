package main

import (
	"context"
	"net"
	"os/signal"
	"syscall"

	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"google.golang.org/grpc"

	triggerpb "github.com/konstellation-io/kai-processes/grpc-trigger/proto"
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
	triggerpb.UnimplementedGRPCTriggerServer
}

func NewServer() *server {
	return &server{}
}

func (s *server) ResponseFunc(ctx context.Context, in *triggerpb.Request) (*triggerpb.Response, error) {
	return &triggerpb.Response{
		// TODO
	}, nil
}

func grpcServerRunner(tr *trigger.Runner, sdk sdk.KaiSDK) {
	sdk.Logger.Info("Starting grpc server", "port", 8080)

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		sdk.Logger.Error(err, "Failed to listen")
	}

	bgCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a gRPC server object
	srv := grpc.NewServer()
	triggerpb.RegisterGRPCTriggerServer(srv, NewServer())

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
	srv.GracefulStop()
	sdk.Logger.Info("Server stopped")
}
