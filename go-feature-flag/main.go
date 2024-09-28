package main

import (
	"context"
	"fmt"
	"github.com/kitabisa/service-client/kuncen"
	"github.com/rs/zerolog/log"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/exporter/logsexporter"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"log/slog"
	"time"
)

type CustomRetriever struct {
}

func (c CustomRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	kuncenClient := kuncen.NewKuncenClient(kuncen.KuncenOption{})

	result, errKuncen := kuncenClient.Get(context.Background(), "dhuwit", "test_ff", "1.0.0")
	if errKuncen != nil {
		log.Err(errKuncen).Msgf("error kuncen")
		return nil, errKuncen
	}

	return result, nil
}

func main() {

	//Init ffclient with a HTTP retriever.
	err := ffclient.Init(ffclient.Config{
		PollingInterval: 10 * time.Second,
		LeveledLogger:   slog.Default(),
		Context:         context.Background(),
		Retriever:       &CustomRetriever{},
		DataExporter: ffclient.DataExporter{
			Exporter: &logsexporter.Exporter{
				LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",
			},
		},
	})

	// Check init errors.
	if err != nil {
		log.Fatal().Msgf("error initializing ffclient: %v", err)
	}
	// defer closing ffclient
	defer ffclient.Close()

	for i := 0; i < 100; i++ {
		time.Sleep(1 * time.Second)
		// create users
		user1 := ffcontext.
			NewEvaluationContextBuilder(fmt.Sprintf("aea2fdc1-b9a0-417a-b707-0c9083de68e3_%d", i)).
			Build()

		// user1
		user1HasAccessToNewAdmin, err := ffclient.BoolVariation("my-first-flag", user1, false)
		if err != nil {
			// we log the error, but we still have a meaningful value in user1HasAccessToNewAdmin (the default value).
			log.Printf("something went wrong when getting the flag: %v", err)
		}

		fmt.Printf("User%d  access to new admin: %v\n", i, user1HasAccessToNewAdmin)
	}

}
