package main

import (
	"os"

	"github.com/konstellation-io/kai-processes/gitlab-webhook-trigger/internal/webhook"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/v2/sdk"
)

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializing webhook")
}

func runnerFunc(wh webhook.Webhook) func(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	return func(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
		kaiSDK.Logger.Info("Running webhook handler")

		err := wh.InitWebhook(kaiSDK)
		if err != nil {
			kaiSDK.Logger.Error(err, "error creating webhook")
			os.Exit(1)
		}
	}
}

func main() {
	wh := webhook.NewGitlabWebhook()

	r := runner.NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(runnerFunc(wh))

	r.Run()
}
