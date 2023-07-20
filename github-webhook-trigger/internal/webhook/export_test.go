//go:build unit

package webhook

import (
	"net/http"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

type GithubWebhookTestExporter struct {
}

func NewGithubWebhookTestExporter() *GithubWebhookTestExporter {
	return &GithubWebhookTestExporter{}
}

func (gw *GithubWebhookTestExporter) HandleEventRequest(hook *github.Webhook, githubEvents []github.Event,
	kaiSDK sdk.KaiSDK) func(w http.ResponseWriter, r *http.Request) {
	return handleEventRequest(hook, githubEvents, kaiSDK)
}

func (gw *GithubWebhookTestExporter) GetEventsFromConfig(eventConfig string) ([]github.Event, error) {
	return getEventsFromConfig(eventConfig)
}
