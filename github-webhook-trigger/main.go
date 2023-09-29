package main

import (
	"os"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializing webhook")

	wh := webhook.NewGithubWebhook()

	err := wh.InitWebhook(kaiSDK)
	if err != nil {
		kaiSDK.Logger.Error(err, "error creating webhook")
		os.Exit(1)
	}
}

func runnerFunc(tr *trigger.Runner, sdk sdk.KaiSDK) {
	sdk.Logger.Info("Running webhook handler")
}

func main() {
	r := runner.NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(runnerFunc)

	r.Run()
}
