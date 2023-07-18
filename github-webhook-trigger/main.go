package main

import (
	"os"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk/messaging"
)

func initializer(webhook webhook.Webhook) func(sdk.KaiSDK) {
	return func(kaiSDK sdk.KaiSDK) {
		webhookEvents, err := kaiSDK.CentralizedConfig.GetConfig("webhook_events", messaging.ProcessScope)
		if err != nil {
			kaiSDK.Logger.Error(err, "Error getting webhook_events config")
			os.Exit(1)
		}

		githubSecret, err := kaiSDK.CentralizedConfig.GetConfig("github_secret", messaging.ProcessScope)
		if err != nil {
			kaiSDK.Logger.Error(err, "Error getting github_secret config")
			os.Exit(1)
		}

		webhook.InitWebhook(webhookEvents, githubSecret, kaiSDK)
	}
}

func main() {
	webhook := webhook.NewGithubWebhook()
	r := runner.NewRunner().TriggerRunner().WithInitializer(initializer(webhook))
	r.Run()
}
