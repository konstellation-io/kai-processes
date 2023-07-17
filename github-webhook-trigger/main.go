package main

import (
	"os"
	"strings"

	webhook "github.com/konstellation-io/kai-processes/github-webhook-trigger/internal"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

func initializer(kaiSDK sdk.KaiSDK) {
	webhookEvents, err := kaiSDK.CentralizedConfig.GetConfig("webhook_events")
	if err != nil {
		kaiSDK.Logger.Error(err, "Error getting webhook_events config")
		os.Exit(1)
	}

	events := strings.Split(webhookEvents, ",")

	githubSecret, err := kaiSDK.CentralizedConfig.GetConfig("github_secret")
	if err != nil {
		kaiSDK.Logger.Error(err, "Error getting github_secret config")
		os.Exit(1)
	}

	webhook.InitWebhook(events, githubSecret, kaiSDK)
}

func main() {
	r := runner.NewRunner().TriggerRunner().WithInitializer(initializer)
	r.Run()
}
