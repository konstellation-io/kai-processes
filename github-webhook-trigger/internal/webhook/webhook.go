package webhook

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	path = "/webhooks"
)

const (
	PushEvent             = "push"
	PullRequestEvent      = "pull_request"
	ReleaseEvent          = "release"
	WorkflowRunEvent      = "workflow_run"
	WorkflowDispatchEvent = "workflow_dispatch"
)

//go:generate mockery --name Webhook --output ../mocks --filename webhook_mock.go --structname WebhookMock
type Webhook interface {
	InitWebhook(eventConfig string, githubSecret string, kaiSDK sdk.KaiSDK) error
}

type GithubWebhook struct {
}

func NewGithubWebhook() Webhook {
	return &GithubWebhook{}
}

func (gw *GithubWebhook) InitWebhook(eventConfig, githubSecret string, kaiSDK sdk.KaiSDK) error {
	githubEvents, err := getEventsFromConfig(eventConfig)
	if err != nil {
		return fmt.Errorf("error getting events from config: %w", err)
	}

	parser, err := github.New(github.Options.Secret(githubSecret))
	if err != nil {
		return fmt.Errorf("error creating webhook: %w", err)
	}

	http.HandleFunc(path, gw.handleEventRequest(parser, githubEvents, kaiSDK))

	server := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("error listening and serving: %w", err)
	}

	return nil
}

func (gw *GithubWebhook) handleEventRequest(parser *github.Webhook, githubEvents []github.Event,
	kaiSDK sdk.KaiSDK) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := parser.Parse(r, githubEvents...)
		if err != nil && !errors.Is(err, github.ErrEventNotFound) {
			kaiSDK.Logger.Error(err, "Error parsing webhook")
			return
		}

		switch payload := payload.(type) {
		case github.PushPayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, PushEvent)
		case github.PullRequestPayload:
			err = triggerPipeline(kaiSDK, payload.PullRequest.URL, PullRequestEvent)
		case github.ReleasePayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, ReleaseEvent)
		case github.WorkflowRunPayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, WorkflowRunEvent)
		case github.WorkflowDispatchPayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, WorkflowDispatchEvent)
		}

		if err != nil {
			kaiSDK.Logger.Error(err, "Error triggering pipeline")
			return
		}
	}
}

func getEventsFromConfig(eventConfig string) ([]github.Event, error) {
	events := strings.Split(strings.ReplaceAll(eventConfig, " ", ""), ",")
	totalEvents := map[string]github.Event{} // use map to avoid duplicates

	for _, event := range events {
		switch event {
		case PushEvent:
			totalEvents[event] = github.PushEvent
		case PullRequestEvent:
			totalEvents[event] = github.PullRequestEvent
		case ReleaseEvent:
			totalEvents[event] = github.ReleaseEvent
		case WorkflowRunEvent:
			totalEvents[event] = github.WorkflowRunEvent
		case WorkflowDispatchEvent:
			totalEvents[event] = github.WorkflowDispatchEvent
		default:
			return nil, fmt.Errorf("%q is not a valid event", event)
		}
	}

	totalEventsSlice := []github.Event{}
	for _, event := range totalEvents {
		totalEventsSlice = append(totalEventsSlice, event)
	}

	return totalEventsSlice, nil
}

func triggerPipeline(kaiSDK sdk.KaiSDK, eventURL, event string) error {
	requestID := uuid.New().String()
	kaiSDK.Logger.Info("Github webhook triggered, new message sent", "requestID", requestID)

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
