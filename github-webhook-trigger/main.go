package main

import (
	"os"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

func initializer(wh webhook.Webhook) func(kaiSDK sdk.KaiSDK) {
	return func(kaiSDK sdk.KaiSDK) {
		kaiSDK.Logger.Info("Initializing webhook")

		err := wh.InitWebhook(kaiSDK)
		if err != nil {
			kaiSDK.Logger.Error(err, "error creating webhook")
			os.Exit(1)
		}
	}
}

func runnerFunc(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Running webhook handler")
}

func main() {
	wh := webhook.NewGithubWebhook()

	r := runner.NewRunner().
		TriggerRunner().
		WithInitializer(initializer(wh)).
		WithRunner(runnerFunc)

	r.Run()
}
