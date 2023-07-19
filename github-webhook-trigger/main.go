package main

import (
	"os"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
)

var (
	webhookEvents string //nolint: gochecknoglobals
	githubSecret  string //nolint: gochecknoglobals
)

func initializer(kaiSDK sdk.KaiSDK) {
	var err error

	webhookEvents, err = kaiSDK.CentralizedConfig.GetConfig("webhook_events", messaging.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, "Error getting webhook_events config")
		os.Exit(1)
	}

	githubSecret, err = kaiSDK.CentralizedConfig.GetConfig("github_secret", messaging.ProcessScope)
	if err != nil {
		kaiSDK.Logger.Error(err, "Error getting github_secret config")
		os.Exit(1)
	}
}

func runnerFunc(wh webhook.Webhook) trigger.RunnerFunc {
	return func(tr *trigger.Runner, sdk sdk.KaiSDK) {
		err := wh.InitWebhook(webhookEvents, githubSecret, sdk)
		if err != nil {
			sdk.Logger.Error(err, "error creating webhook")
			os.Exit(1)
		}
	}
}

func main() {
	wh := webhook.NewGithubWebhook()
	r := runner.NewRunner().TriggerRunner().WithInitializer(initializer).WithRunner(runnerFunc(wh))
	r.Run()
}
