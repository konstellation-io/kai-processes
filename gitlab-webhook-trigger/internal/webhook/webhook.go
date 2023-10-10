package webhook

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	centralizedconfiguration "github.com/konstellation-io/kai-sdk/go-sdk/sdk/centralized-configuration"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	path = "/webhook-gitlab"
)

const (
	PushEvent         = "push"
	MergeRequestEvent = "merge_request"
	CommentEvent      = "comment"
	TagEvent          = "tag"
)

//go:generate mockery --name Webhook --output ../mocks --filename webhook_mock.go --structname WebhookMock
type Webhook interface {
	InitWebhook(kaiSDK sdk.KaiSDK) error
}

type GitlabWebhook struct {
}

func NewGitlabWebhook() Webhook {
	return &GitlabWebhook{}
}

func (gw *GitlabWebhook) InitWebhook(kaiSDK sdk.KaiSDK) error {
	eventConfig, gitlabSecret, err := getConfig(kaiSDK)
	if err != nil {
		return err
	}

	gitlabEvents, err := getEventsFromConfig(eventConfig)
	if err != nil {
		return GettingEventsFromConfigError(err)
	}

	options := make([]gitlab.Option, 0, 1)
	if gitlabSecret != "" {
		options = append(options, gitlab.Options.Secret(gitlabSecret))
	}

	parser, err := gitlab.New(options...)
	if err != nil {
		return CreatingWebhookError(err)
	}

	http.HandleFunc(path, handleEventRequest(parser, gitlabEvents, kaiSDK))

	srv := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: 5 * time.Second,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		kaiSDK.Logger.Info("Server listed", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			signalCh <- syscall.SIGTERM
		}
	}()

	sig := <-signalCh
	kaiSDK.Logger.Info("Received signal", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		kaiSDK.Logger.Error(err, "Server shutdown failed")
	}

	kaiSDK.Logger.Info("Server shutdown gracefully")

	return nil
}

func getConfig(kaiSDK sdk.KaiSDK) (webhookEvents, gitlabSecret string, err error) {
	webhookEvents, err = kaiSDK.CentralizedConfig.GetConfig("webhook_events", messaging.ProcessScope)
	if err != nil {
		return "", "", err
	}

	gitlabSecret, err = kaiSDK.CentralizedConfig.GetConfig("gitlab_secret", messaging.ProcessScope)
	if err != nil && !errors.Is(err, centralizedconfiguration.ErrKeyNotFound) {
		return "", "", err
	}

	return webhookEvents, gitlabSecret, nil
}

func handleEventRequest(
	parser *gitlab.Webhook,
	gitlabEvents []gitlab.Event,
	kaiSDK sdk.KaiSDK,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := parser.Parse(r, gitlabEvents...)
		if err != nil && !errors.Is(err, gitlab.ErrEventNotFound) {
			kaiSDK.Logger.Error(err, "Error parsing webhook")
			return
		}

		switch payload := payload.(type) {
		case gitlab.PushEventPayload:
			err = triggerPipeline(payload.Repository.URL, string(gitlab.PushEvents), kaiSDK)
		case gitlab.MergeRequestEventPayload:
			err = triggerPipeline(payload.Repository.URL, string(gitlab.MergeRequestEvents), kaiSDK)
		case gitlab.CommentEventPayload:
			err = triggerPipeline(payload.Repository.URL, string(gitlab.CommentEvents), kaiSDK)
		case gitlab.TagEventPayload:
			err = triggerPipeline(payload.Repository.URL, string(gitlab.TagEvents), kaiSDK)
		default:
			err = ErrEventNotSupported
		}

		if err != nil {
			kaiSDK.Logger.Error(err, "Error triggering pipeline")
			return
		}
	}
}

func getEventsFromConfig(eventConfig string) ([]gitlab.Event, error) {
	events := strings.Split(strings.ReplaceAll(eventConfig, " ", ""), ",")
	totalEvents := map[string]gitlab.Event{} // use map to avoid duplicates

	for _, event := range events {
		switch event {
		case PushEvent:
			totalEvents[event] = gitlab.PushEvents
		case MergeRequestEvent:
			totalEvents[event] = gitlab.MergeRequestEvents
		case CommentEvent:
			totalEvents[event] = gitlab.CommentEvents
		case TagEvent:
			totalEvents[event] = gitlab.TagEvents
		default:
			return nil, NotValidEventError(event)
		}
	}

	totalEventsSlice := []gitlab.Event{}
	for _, event := range totalEvents {
		totalEventsSlice = append(totalEventsSlice, event)
	}

	return totalEventsSlice, nil
}

func triggerPipeline(eventURL, event string, kaiSDK sdk.KaiSDK) error {
	requestID := uuid.New().String()
	kaiSDK.Logger.Info("Gitlab webhook triggered, new message sent", "requestID", requestID)

	m, err := structpb.NewValue(map[string]interface{}{
		"eventUrl": eventURL,
		"event":    event,
	})
	if err != nil {
		return err
	}

	err = kaiSDK.Messaging.SendOutputWithRequestID(m, requestID)
	if err != nil {
		return err
	}

	return nil
}
