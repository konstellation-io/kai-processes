package main

import (
	cronjob "algo/kai-processes/cronjob-trigger/internal/cronjob"
	"os"

	"github.com/konstellation-io/kai-sdk/go-sdk/runner"
	"github.com/konstellation-io/kai-sdk/go-sdk/runner/trigger"
	"github.com/konstellation-io/kai-sdk/go-sdk/sdk"
)

func initializer(kaiSDK sdk.KaiSDK) {
	kaiSDK.Logger.Info("Initializing cronjob")
}

func runnerFunc(cr cronjob.Cronjob) func(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
	return func(tr *trigger.Runner, kaiSDK sdk.KaiSDK) {
		kaiSDK.Logger.Info("Running cronjob handler")

		err := cr.InitCronjob(kaiSDK)
		if err != nil {
			kaiSDK.Logger.Error(err, "error creating cronjob")
			os.Exit(1)
		}
	}
}

func main() {
	cr := cronjob.NewCronjob()

	r := runner.NewRunner().
		TriggerRunner().
		WithInitializer(initializer).
		WithRunner(runnerFunc(cr))

	r.Run()
}
