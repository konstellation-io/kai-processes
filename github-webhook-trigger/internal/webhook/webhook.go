package webhook

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/google/uuid"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	path = "/webhooks"
)

const (
	PullRequestEvent      = "pull_request"
	PushEvent             = "push"
	ReleaseEvent          = "release"
	WorkflowDispatchEvent = "workflow_dispatch"
	WorkflowRunEvent      = "workflow_run"
)

//go:generate mockery --name Webhook --output ../mocks --filename webhook_mock.go --structname WebhookMock
type Webhook interface {
	InitWebhook(kaiSDK sdk.KaiSDK) error
}

type GithubWebhook struct {
}

func NewGithubWebhook() Webhook {
	return &GithubWebhook{}
}

func (gw *GithubWebhook) InitWebhook(kaiSDK sdk.KaiSDK) error {
	eventConfig, githubSecret, err := getConfig(kaiSDK)
	if err != nil {
		return err
	}

	githubEvents, err := getEventsFromConfig(eventConfig)
	if err != nil {
		return GettingEventsFromConfigError(err)
	}

	parser, err := github.New(github.Options.Secret(githubSecret))
	if err != nil {
		return CreatingWebhookError(err)
	}

	http.HandleFunc(path, handleEventRequest(parser, githubEvents, kaiSDK))

	server := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: 5 * time.Second,
	}
	defer server.Close()

	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	} else if err != nil {
		return ServerError(err)
	}

	return nil
}

func getConfig(kaiSDK sdk.KaiSDK) (string, string, error) {
	webhookEvents, err := kaiSDK.CentralizedConfig.GetConfig("webhook_events", messaging.ProcessScope)
	if err != nil {
		return "", "", err
	}

	githubSecret, err := kaiSDK.CentralizedConfig.GetConfig("github_secret", messaging.ProcessScope)
	if err != nil {
		return "", "", err
	}

	return webhookEvents, githubSecret, nil
}

func handleEventRequest(
	parser *github.Webhook,
	githubEvents []github.Event,
	kaiSDK sdk.KaiSDK,
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := parser.Parse(r, githubEvents...)
		if err != nil && !errors.Is(err, github.ErrEventNotFound) {
			kaiSDK.Logger.Error(err, "Error parsing webhook")
			return
		}

		switch payload := payload.(type) {
		case github.PullRequestPayload:
			err = triggerPipeline(payload.PullRequest.URL, PullRequestEvent, kaiSDK)
		case github.PushPayload:
			err = triggerPipeline(payload.Repository.URL, PushEvent, kaiSDK)
		case github.ReleasePayload:
			err = triggerPipeline(payload.Repository.URL, ReleaseEvent, kaiSDK)
		case github.WorkflowDispatchPayload:
			err = triggerPipeline(payload.Repository.URL, WorkflowDispatchEvent, kaiSDK)
		case github.WorkflowRunPayload:
			err = triggerPipeline(payload.Repository.URL, WorkflowRunEvent, kaiSDK)
		default:
			err = ErrEventNotSupported
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
		case PullRequestEvent:
			totalEvents[event] = github.PullRequestEvent
		case PushEvent:
			totalEvents[event] = github.PushEvent
		case ReleaseEvent:
			totalEvents[event] = github.ReleaseEvent
		case WorkflowDispatchEvent:
			totalEvents[event] = github.WorkflowDispatchEvent
		case WorkflowRunEvent:
			totalEvents[event] = github.WorkflowRunEvent
		default:
			return nil, NotValidEventError(event)
		}
	}

	totalEventsSlice := []github.Event{}
	for _, event := range totalEvents {
		totalEventsSlice = append(totalEventsSlice, event)
	}

	return totalEventsSlice, nil
}

func triggerPipeline(eventURL, event string, kaiSDK sdk.KaiSDK) error {
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
