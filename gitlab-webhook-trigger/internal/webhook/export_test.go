//go:build unit

package webhook

import (
	"net/http"

	"github.com/go-playground/webhooks/v6/gitlab"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

type GitlabWebhookTestExporter struct{}

func NewGitlabWebhookTestExporter() *GitlabWebhookTestExporter {
	return &GitlabWebhookTestExporter{}
}

func (gw *GitlabWebhookTestExporter) HandleEventRequest(
	hook *gitlab.Webhook,
	gitlabEvents []gitlab.Event,
	kaiSDK sdk.KaiSDK,
) func(w http.ResponseWriter, r *http.Request) {
	return handleEventRequest(hook, gitlabEvents, kaiSDK)
}

func (gw *GitlabWebhookTestExporter) GetEventsFromConfig(eventConfig string) ([]gitlab.Event, error) {
	return getEventsFromConfig(eventConfig)
}
