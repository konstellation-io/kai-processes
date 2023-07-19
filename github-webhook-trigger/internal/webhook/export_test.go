package webhook

import (
	"net/http"

	"github.com/go-playground/webhooks/v6/github"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

func NewTestGithubWebhook() *GithubWebhook {
	return &GithubWebhook{}
}

func (gw *GithubWebhook) HandleEventRequest(hook *github.Webhook, githubEvents []github.Event,
	kaiSDK sdk.KaiSDK) func(w http.ResponseWriter, r *http.Request) {
	return gw.handleEventRequest(hook, githubEvents, kaiSDK)
}

func (gw *GithubWebhook) GetEventsFromConfig(eventConfig string) ([]github.Event, error) {
	return getEventsFromConfig(eventConfig)
}
