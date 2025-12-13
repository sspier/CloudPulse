package main

import (
	"context"
	"log"
	"os"
	"time"

	awsLambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/sspier/cloudpulse/internal/store"
)

// bootstraps the Lambda environment and starts the handler to write data and run probes
func main() {
	// initialize store based on environment
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	targetsTable := os.Getenv("TABLE_NAME_TARGETS")
	resultsTable := os.Getenv("TABLE_NAME_RESULTS")

	if targetsTable == "" || resultsTable == "" {
		log.Fatal("TABLE_NAME_TARGETS and TABLE_NAME_RESULTS must be set")
	}

	// use context.Background() for init, but handler will provide its own context
	dynamoDBStore, err := store.NewDynamoDBStore(context.Background(), awsRegion, targetsTable, resultsTable)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	handler := &Handler{
		store: dynamoDBStore,
	}

	// check if running in Lambda
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") != "" {
		awsLambda.Start(handler.HandleRequest)
	} else {
		log.Println("running in local mode (poll loop)")
		// run once immediately
		if _, err := handler.HandleRequest(context.Background()); err != nil {
			log.Printf("initial check failed: %v", err)
		}

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := handler.HandleRequest(context.Background()); err != nil {
				log.Printf("check failed: %v", err)
			}
		}
	}
}
