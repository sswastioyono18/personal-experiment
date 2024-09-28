package main

import (
	"context"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
	"log"
	"log/slog"
	"time"

	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func main() {
	// Before running this code please check the flag.yaml file
	// You can update the dates of the steps in the rollout to see it working.

	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		LeveledLogger:   slog.Default(),
		Context:         context.Background(),
		Retriever: &githubretriever.Retriever{
			RepositorySlug: "sswastioyono18/test-public-repo",
			Branch:         "main",
			FilePath:       "ff.yaml",
			//FilePath:       "ff_experiment.yaml",
			Timeout: 3 * time.Second,
		},
	})
	// Check init errors.
	if err != nil {
		log.Fatal(err)
	}
	// defer closing ffclient
	defer ffclient.Close()

	// create users

	// Call multiple time the same flag to see the change in time.
	for i := 0; i < 100; i++ {
		user := ffcontext.NewEvaluationContextBuilder(fmt.Sprintf("user-id:%d", i)).
			AddCustom("beta", "true").
			Build()

		time.Sleep(1 * time.Second)
		fmt.Println(ffclient.BoolVariation("progressive-flag", user, false))
	}

	//for i := 0; i < 100; i++ {
	//	user := ffcontext.NewEvaluationContextBuilder(fmt.Sprintf("user-id:%d", i)).
	//		AddCustom("beta", "true").
	//		Build()
	//
	//	time.Sleep(1 * time.Second)
	//	fmt.Println(ffclient.StringVariation("experimentation-flag", user, "error"))
	//}
}
