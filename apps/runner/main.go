package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sspier/cloudpulse/internal/store"
)

func main() {
	// Initialize store based on environment
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	targetsTable := os.Getenv("TABLE_NAME_TARGETS")
	resultsTable := os.Getenv("TABLE_NAME_RESULTS")

	if targetsTable == "" || resultsTable == "" {
		log.Fatal("TABLE_NAME_TARGETS and TABLE_NAME_RESULTS must be set")
	}

	// We use context.Background() for init, but handler will provide its own context
	st, err := store.NewDynamoDBStore(context.Background(), region, targetsTable, resultsTable)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}

	h := &Handler{
		store: st,
	}

	lambda.Start(h.HandleRequest)
}
