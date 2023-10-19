package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
	runner.
		NewRunner().
		TriggerRunner().
		WithInitializer(func(ctx sdk.KaiSDK) {
			ctx.Logger.Info("Initializer")
		}).
		WithRunner(restServerRunner).
		WithFinalizer(func(ctx sdk.KaiSDK) {
			ctx.Logger.Info("Finalizer")
		}).
		Run()
}

func restServerRunner(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Starting http server", "port", 8080)

	bgCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r := gin.Default()
	r.POST("/trigger", responseHandler(kaiSDK, tr.GetResponseChannel))
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			kaiSDK.Logger.Error(err, "Error running http server")
		}
	}()

	<-bgCtx.Done()
	stop()
	kaiSDK.Logger.Info("Shutting down server...")

	// The kaiSDK is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(bgCtx); err != nil {
		kaiSDK.Logger.Error(err, "Error shutting down server")
	}

	kaiSDK.Logger.Info("Server stopped")
}

func responseHandler(kaiSDK sdk.KaiSDK, getResponseChannel func(requestID string) <-chan *anypb.Any) func(c *gin.Context) {
	return func(c *gin.Context) {

		reqID := uuid.New().String()
		kaiSDK.Logger.Info("REST triggered, sending message", "requestID", reqID)

		// jsonData, err := io.ReadAll(c.Request.Body)
		// if err != nil {
		// 	kaiSDK.Logger.Error(err, "Error reading request body")
		// 	return
		// }

		var requestData struct {
			Param1 string `json:"param1"`
			Param2 string `json:"param2"`
			Param3 string `json:"param3"`
		}

		err := c.ShouldBindJSON(&requestData)
		if err != nil {
			kaiSDK.Logger.Error(err, "Error binding request body")
			return
		}

		m, err := structpb.NewValue(map[string]interface{}{
			"param1": requestData.Param1,
			"param2": requestData.Param2,
			"param3": requestData.Param3,
		})
		if err != nil {
			kaiSDK.Logger.Error(err, "error creating response")
			return
		}

		err = kaiSDK.Messaging.SendOutputWithRequestID(m, reqID)
		if err != nil {
			kaiSDK.Logger.Error(err, "Error sending output")
			return
		}

		responseChannel := getResponseChannel(reqID)
		response := <-responseChannel

		kaiSDK.Logger.Info("response recieved", "response", response)

		var respData struct {
			StatusCode string `json:"status_code"`
			Message    string `json:"message"`
		}

		responsePb := new(structpb.Value)
		if err := response.UnmarshalTo(responsePb); err != nil {
			kaiSDK.Logger.Error(err, "error while creating Value from Any")
			return
		}
		response.UnmarshalTo(responsePb)

		responsePbJSON, err := responsePb.MarshalJSON()
		if err != nil {
			kaiSDK.Logger.Error(err, "error marshalling response")
			return
		}

		kaiSDK.Logger.Info("json bytes", "json", string(responsePbJSON))

		err = json.Unmarshal(responsePbJSON, &respData)
		if err != nil {
			kaiSDK.Logger.Error(err, "error unmarshalling response")
			return
		}

		httpCode, err := strconv.Atoi(respData.StatusCode)
		if err != nil {
			kaiSDK.Logger.Error(err, "error converting status code to int")
			return
		}

		c.JSON(httpCode, gin.H{
			"message": respData.Message,
		})
	}
}
