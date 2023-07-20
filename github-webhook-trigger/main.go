package main

import (
	"os"

	"github.com/konstellation-io/kai-processes/github-webhook-trigger/internal/webhook"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

func runnerFunc(wh webhook.Webhook) trigger.RunnerFunc {
	return func(tr *trigger.Runner, sdk sdk.KaiSDK) {
		err := wh.InitWebhook(sdk)
		if err != nil {
			sdk.Logger.Error(err, "error creating webhook")
			os.Exit(1)
		}
	}
}

func main() {
	wh := webhook.NewGithubWebhook()
	r := runner.NewRunner().TriggerRunner().WithRunner(runnerFunc(wh))
	r.Run()
}
