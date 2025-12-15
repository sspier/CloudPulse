package main

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sspier/cloudpulse/internal/probe"
	"github.com/sspier/cloudpulse/internal/store"
)

// TestIntegrationWithLocalDynamoDB verifies the full flow using the local DynamoDB container
// pre-requisite: DynamoDB Local must be running on localhost:8000 (e.g. via kubectl port-forward)
func TestIntegrationWithLocalDynamoDB(t *testing.T) {
	ctx := context.Background()
	const awsRegion = "us-east-1"

	// configure the AWS SDK to point to local DynamoDB (localhost:8000)
	// use a custom resolver because standard AWS regions don't map to localhost
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           "http://localhost:8000",
			SigningRegion: awsRegion,
		}, nil
	})

	// load AWS config with local settings and dummy credentials
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsRegion),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: "dummy", SecretAccessKey: "dummy"}, nil
		})),
	)
	if err != nil {
		t.Skipf("Skipping test: could not load aws config: %v", err)
	}

	// verify that the Docker container is up and tables are created
	dynamoDBClient := dynamodb.NewFromConfig(awsConfig)
	listTablesOutput, err := dynamoDBClient.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		t.Skipf("Skipping test: local dynamodb not reachable (is docker running?): %v", err)
	}

	// check if specific testing tables exist
	foundTargets := false
	for _, name := range listTablesOutput.TableNames {
		if name == "cloudpulse-targets-local" {
			foundTargets = true
			break
		}
	}
	if !foundTargets {
		t.Skip("Skipping test: 'cloudpulse-targets-local' table not found. Waiting for docker-compose init?")
	}

	// initialize the Application Store
	// since the fields in DynamoDBStore are unexported, we cannot initialize it directly
	// as a struct literal in this test package (which is separate from internal/store).
	// instead, we can use the constructor NewDynamoDBStore.
	//
	// NewDynamoDBStore initializes its own client using the default AWS config loading.
	// to make that internal client point to our local DynamoDB, we need to set the
	// AWS_ENDPOINT_URL environment variable, which the AWS SDK v2 respects.
	os.Setenv("AWS_ENDPOINT_URL", "http://localhost:8000")
	os.Setenv("AWS_REGION", awsRegion)
	os.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")

	// load these env vars and point to localhost
	dynamoDBStore, err := store.NewDynamoDBStore(ctx, awsRegion, "cloudpulse-targets-local", "cloudpulse-probe-results-local")
	if err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// test scenario: add a target -> check it -> store result

	// add a target
	t.Log("Adding target...")
	targetURL := "https://example.com"
	target, err := dynamoDBStore.AddTarget(ctx, "Integration Test Target", targetURL)
	if err != nil {
		t.Fatalf("Failed to add target: %v", err)
	}

	// run the probe (simulating the runner)
	t.Log("Running probe...")
	result := probe.Check(ctx, target)
	if result.TargetID != target.ID {
		t.Errorf("Result mismatched target ID: got %v, want %v", result.TargetID, target.ID)
	}

	// store the result
	t.Log("Storing result...")
	if err := dynamoDBStore.AddResult(ctx, result); err != nil {
		t.Fatalf("Failed to add result to store: %v", err)
	}

	// verify persistence
	results, err := dynamoDBStore.ResultsForTarget(ctx, target.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve results: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least 1 result in database, found 0")
	}

	t.Logf("Success! Verified %d result(s) stored for target %s", len(results), target.ID)
}
