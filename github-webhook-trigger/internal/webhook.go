package webhook

import (
	"errors"
	"net/http"
	"os"
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
	pushEvent             = "push"
	pullRequestEvent      = "pull_request"
	releaseEvent          = "release"
	workflowRunEvent      = "workflow_run"
	workflowDispatchEvent = "workflow_dispatch"
)

//go:generate mockery --name Webhook --output ../mocks --filename webhook_mock.go --structname WebhookMock

type Webhook interface {
	InitWebhook(events []string, githubSecret string, kaiSDK sdk.KaiSDK)
}

type GithubWebhook struct {
}

func NewGithubWebhook() Webhook {
	return &GithubWebhook{}
}

func (gw *GithubWebhook) InitWebhook(events []string, githubSecret string, kaiSDK sdk.KaiSDK) {
	githubEvents := getEventsFromConfig(events)

	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		hook, err := github.New(github.Options.Secret(githubSecret))
		if err != nil {
			kaiSDK.Logger.Error(err, "Error creating webhook")
			os.Exit(1)
		}

		payload, err := hook.Parse(r, githubEvents...)
		if err != nil && !errors.Is(err, github.ErrEventNotFound) {
			kaiSDK.Logger.Error(err, "Error parsing webhook")
			return
		}

		switch payload := payload.(type) {
		case github.PushPayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, pushEvent)
		case github.PullRequestPayload:
			err = triggerPipeline(kaiSDK, payload.PullRequest.URL, pullRequestEvent)
		case github.ReleasePayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, releaseEvent)
		case github.WorkflowRunPayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, workflowRunEvent)
		case github.WorkflowDispatchPayload:
			err = triggerPipeline(kaiSDK, payload.Repository.URL, workflowDispatchEvent)
		}

		if err != nil {
			kaiSDK.Logger.Error(err, "Error triggering pipeline")
			return
		}
	})

	server := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		kaiSDK.Logger.Error(err, "Error listening and serving")
		os.Exit(1)
	}
}

func getEventsFromConfig(eventConfig []string) []github.Event {
	totalEvents := map[string]github.Event{} // use map to avoid duplicates

	for _, event := range eventConfig {
		switch event {
		case pushEvent:
			totalEvents[event] = github.PushEvent
		case pullRequestEvent:
			totalEvents[event] = github.PullRequestEvent
		case releaseEvent:
			totalEvents[event] = github.ReleaseEvent
		case workflowRunEvent:
			totalEvents[event] = github.WorkflowRunEvent
		case workflowDispatchEvent:
			totalEvents[event] = github.WorkflowDispatchEvent
		}
	}

	totalEventsSlice := []github.Event{}
	for _, event := range totalEvents {
		totalEventsSlice = append(totalEventsSlice, event)
	}

	return totalEventsSlice
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
