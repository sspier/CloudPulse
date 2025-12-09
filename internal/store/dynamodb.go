package store

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sspier/cloudpulse/internal/model"
)

type DynamoDBStore struct {
	client       *dynamodb.Client
	targetsTable string
	resultsTable string
}

func NewDynamoDBStore(ctx context.Context, region, targetsTable, resultsTable string) (*DynamoDBStore, error) {
	// load the default config for the region
	awsConfig, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	// create a new DynamoDB client from the config
	return &DynamoDBStore{
		client:       dynamodb.NewFromConfig(awsConfig),
		targetsTable: targetsTable,
		resultsTable: resultsTable,
	}, nil
}

// AddTarget adds a target to the targets table
func (dynamoDBStore *DynamoDBStore) AddTarget(ctx context.Context, name, url string) (model.Target, error) {
	// use the timestamp as ID, similar to the in-memory store
	targetId := strconv.FormatInt(timeNow().UnixNano(), 10)

	// create the target
	target := model.Target{
		ID:   targetId,
		Name: name,
		URL:  url,
	}

	// convert the target to a map for DynamoDB
	attributeValue, err := attributevalue.MarshalMap(target)
	// if marshaling fails, return an error
	if err != nil {
		return model.Target{}, fmt.Errorf("failed to marshal target: %w", err)
	}

	// insert the target into the targets table
	_, err = dynamoDBStore.client.PutItem(ctx, &dynamodb.PutItemInput{
		// the name of the table
		TableName: aws.String(dynamoDBStore.targetsTable),
		// the item to insert
		Item: attributeValue,
	})

	// if inserting fails, return an error
	if err != nil {
		return model.Target{}, fmt.Errorf("failed to put item to dynamodb: %w", err)
	}

	return target, nil
}

// ListTargets scans the targets table and returns all targets
func (dynamoDBStore *DynamoDBStore) ListTargets(ctx context.Context) ([]model.Target, error) {
	var targets []model.Target
	// create a paginator to scan the targets table
	paginator := dynamodb.NewScanPaginator(dynamoDBStore.client, &dynamodb.ScanInput{
		TableName: aws.String(dynamoDBStore.targetsTable),
	})

	// iterate through the pages of the paginator
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to scan targets: %w", err)
		}

		var pageTargets []model.Target
		// unmarshal the page of targets into a list of targets
		err = attributevalue.UnmarshalListOfMaps(page.Items, &pageTargets)
		// if unmarshaling fails, return an error
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal targets: %w", err)
		}
		// append the page of targets to the list of targets
		targets = append(targets, pageTargets...)
	}

	return targets, nil
}

// AddResult adds a result to the results table
func (dynamoDBStore *DynamoDBStore) AddResult(ctx context.Context, result model.Result) error {
	// convert the result to a map for DynamoDB
	attributeValue, err := attributevalue.MarshalMap(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// add TTL if needed
	attributeValue["ttl"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(timeNow().Add(30*24*time.Hour).Unix(), 10)}

	// insert the result into the results table
	_, err = dynamoDBStore.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(dynamoDBStore.resultsTable),
		Item:      attributeValue,
	})

	return err
}

// ResultsForTarget queries the results table for a specific target
func (dynamoDBStore *DynamoDBStore) ResultsForTarget(ctx context.Context, targetID string) ([]model.Result, error) {
	// query the results table for a specific target
	// awsQueryInput is a pointer to a QueryInput struct from the AWS SDK
	awsQueryInput := &dynamodb.QueryInput{
		TableName:              aws.String(dynamoDBStore.resultsTable),
		KeyConditionExpression: aws.String("target_id = :tid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: targetID},
		},
		ScanIndexForward: aws.Bool(false), // descending order
		Limit:            aws.Int32(100),  // limit to last 100 results for now
	}

	// query the results table
	// output is a pointer to a QueryOutput struct from the AWS SDK
	awsQueryOutput, err := dynamoDBStore.client.Query(ctx, awsQueryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query results: %w", err)
	}

	var results []model.Result
	// unmarshal the results into a list of results
	err = attributevalue.UnmarshalListOfMaps(awsQueryOutput.Items, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return results, nil
}

// LatestResults gets the latest result for each target
// for now, we fetch all targets and query one latest result for each
func (dynamoDBStore *DynamoDBStore) LatestResults(ctx context.Context) ([]model.Result, error) {
	// list all listOfTargets
	listOfTargets, err := dynamoDBStore.ListTargets(ctx)
	if err != nil {
		return nil, err
	}

	var latestResults []model.Result
	for _, target := range listOfTargets {
		// query just 1 item
		awsQueryInput := &dynamodb.QueryInput{
			TableName:              aws.String(dynamoDBStore.resultsTable),
			KeyConditionExpression: aws.String("target_id = :tid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tid": &types.AttributeValueMemberS{Value: target.ID},
			},
			ScanIndexForward: aws.Bool(false),
			Limit:            aws.Int32(1),
		}

		awsQueryOutput, err := dynamoDBStore.client.Query(ctx, awsQueryInput)
		if err != nil {
			// LOG but continue?
			continue
		}

		if len(awsQueryOutput.Items) > 0 {
			var result model.Result
			_ = attributevalue.UnmarshalMap(awsQueryOutput.Items[0], &result)
			latestResults = append(latestResults, result)
		}
	}
	return latestResults, nil
}

// Helper for mocking time in tests if needed, though we just use time.Now() here
func timeNow() time.Time {
	return time.Now().UTC()
}

// Fix unused sort import
var _ = sort.Interface(nil)
