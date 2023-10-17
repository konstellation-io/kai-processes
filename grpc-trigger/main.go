package main

import (
	"context"
	"encoding/json"
	"net"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

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
	tr     *trigger.Runner
	kaiSDK sdk.KaiSDK
}

func NewServer(tr *trigger.Runner, kaiSDK sdk.KaiSDK) *server {
	return &server{
		tr:     tr,
		kaiSDK: kaiSDK,
	}
}

func (s *server) Trigger(ctx context.Context, req *triggerpb.Request) (*triggerpb.Response, error) {
	reqID := uuid.New().String()
	s.kaiSDK.Logger.Info("GRPC triggered, new message sent", "requestID", reqID)

	m, err := structpb.NewValue(map[string]interface{}{
		"param1": req.GetParam1(),
		"param2": req.GetParam2(),
		"param3": req.GetParam3(),
	})
	if err != nil {
		s.kaiSDK.Logger.Error(err, "error creating response")
		return nil, err
	}

	err = s.kaiSDK.Messaging.SendOutputWithRequestID(m, reqID)
	if err != nil {
		s.kaiSDK.Logger.Error(err, "Error sending output")
		return nil, err
	}

	responseChannel := s.tr.GetResponseChannel(reqID)
	response := <-responseChannel

	var respData struct {
		StatusCode string `json:"status_code"`
		Message    string `json:"message"`
	}

	err = json.Unmarshal(response.GetValue(), &respData)
	if err != nil {
		s.kaiSDK.Logger.Error(err, "error unmarshalling response")
		return nil, err
	}

	return &triggerpb.Response{
		StatusCode: respData.StatusCode,
		Message:    respData.Message,
	}, nil
}

func grpcServerRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Starting grpc server", "port", 8080)

	// Create a listener on TCP port
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		kaiSDK.Logger.Error(err, "Failed to listen")
	}

	bgCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create a gRPC server object
	srv := grpc.NewServer()
	triggerpb.RegisterGRPCTriggerServer(srv, NewServer(tr, kaiSDK))

	// Serve gRPC Server
	kaiSDK.Logger.Info("Serving gRPC on 0.0.0.0:8080")
	go func() {
		if err := srv.Serve(lis); err != nil {
			kaiSDK.Logger.Error(err, "Failed to serve")
		}
	}()

	<-bgCtx.Done()
	stop()

	kaiSDK.Logger.Info("Shutting down server...")
	srv.GracefulStop()
	kaiSDK.Logger.Info("Server stopped")
}
