// Package services contains business logic and external service integrations.
package services

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"steverhoton-labor-lines/lambda/models"
)

// DynamoDBService defines the interface for DynamoDB operations.
type DynamoDBService interface {
	CreateLaborLine(ctx context.Context, laborLine *models.LaborLine) error
	GetLaborLine(ctx context.Context, input models.GetLaborLineInput) (*models.LaborLine, error)
	UpdateLaborLine(ctx context.Context, laborLine *models.LaborLine) error
	DeleteLaborLine(ctx context.Context, input models.DeleteLaborLineInput) error
	ListLaborLines(ctx context.Context, input models.ListLaborLinesInput) ([]*models.LaborLine, error)
}

// DynamoDBClient defines the interface for DynamoDB client operations we use.
type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
}

// dynamoDBService implements DynamoDBService.
type dynamoDBService struct {
	client    DynamoDBClient
	tableName string
}

// NewDynamoDBService creates a new DynamoDB service instance.
func NewDynamoDBService(client DynamoDBClient, tableName string) DynamoDBService {
	return &dynamoDBService{
		client:    client,
		tableName: tableName,
	}
}

// CreateLaborLine creates a new labor line in DynamoDB.
func (s *dynamoDBService) CreateLaborLine(ctx context.Context, laborLine *models.LaborLine) error {
	item, err := attributevalue.MarshalMap(laborLine)
	if err != nil {
		return fmt.Errorf("marshaling labor line: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}

	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("creating labor line in DynamoDB: %w", err)
	}

	return nil
}

// GetLaborLine retrieves a labor line from DynamoDB.
func (s *dynamoDBService) GetLaborLine(ctx context.Context, input models.GetLaborLineInput) (*models.LaborLine, error) {
	pk := input.AccountID
	sk := input.TaskID + "#" + input.LaborLineID

	getInput := &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
		},
	}

	result, err := s.client.GetItem(ctx, getInput)
	if err != nil {
		return nil, fmt.Errorf("getting labor line from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, nil // Not found
	}

	var laborLine models.LaborLine
	err = attributevalue.UnmarshalMap(result.Item, &laborLine)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling labor line: %w", err)
	}

	// Don't return soft-deleted items
	if laborLine.IsDeleted() {
		return nil, nil
	}

	return &laborLine, nil
}

// UpdateLaborLine updates an existing labor line in DynamoDB.
func (s *dynamoDBService) UpdateLaborLine(ctx context.Context, laborLine *models.LaborLine) error {
	// First, get the existing item to preserve createdAt and ensure it exists
	existing, err := s.GetLaborLine(ctx, models.GetLaborLineInput{
		AccountID:   laborLine.AccountID,
		TaskID:      laborLine.TaskID,
		LaborLineID: laborLine.LaborLineID,
	})
	if err != nil {
		return fmt.Errorf("checking existing labor line: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("labor line not found")
	}

	// Preserve the original createdAt timestamp
	laborLine.CreatedAt = existing.CreatedAt

	item, err := attributevalue.MarshalMap(laborLine)
	if err != nil {
		return fmt.Errorf("marshaling labor line: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK) AND attribute_not_exists(deletedAt)"),
	}

	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("updating labor line in DynamoDB: %w", err)
	}

	return nil
}

// DeleteLaborLine soft deletes a labor line in DynamoDB.
func (s *dynamoDBService) DeleteLaborLine(ctx context.Context, input models.DeleteLaborLineInput) error {
	// First get the existing item
	existing, err := s.GetLaborLine(ctx, models.GetLaborLineInput{
		AccountID:   input.AccountID,
		TaskID:      input.TaskID,
		LaborLineID: input.LaborLineID,
	})
	if err != nil {
		return fmt.Errorf("checking existing labor line: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("labor line not found")
	}

	// Soft delete the item
	existing.SoftDelete()

	item, err := attributevalue.MarshalMap(existing)
	if err != nil {
		return fmt.Errorf("marshaling labor line for deletion: %w", err)
	}

	updateInput := &dynamodb.PutItemInput{
		TableName:           aws.String(s.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	}

	_, err = s.client.PutItem(ctx, updateInput)
	if err != nil {
		return fmt.Errorf("soft deleting labor line in DynamoDB: %w", err)
	}

	return nil
}

// ListLaborLines retrieves labor lines for an account, optionally filtered by task.
func (s *dynamoDBService) ListLaborLines(ctx context.Context, input models.ListLaborLinesInput) ([]*models.LaborLine, error) {
	var queryInput *dynamodb.QueryInput

	if input.TaskID != "" {
		// Query by specific task
		queryInput = &dynamodb.QueryInput{
			TableName:              aws.String(s.tableName),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :skPrefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":       &types.AttributeValueMemberS{Value: input.AccountID},
				":skPrefix": &types.AttributeValueMemberS{Value: input.TaskID + "#"},
			},
		}
	} else {
		// Query all labor lines for the account
		queryInput = &dynamodb.QueryInput{
			TableName:              aws.String(s.tableName),
			KeyConditionExpression: aws.String("PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: input.AccountID},
			},
		}
	}

	result, err := s.client.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("querying labor lines from DynamoDB: %w", err)
	}

	var laborLines []*models.LaborLine
	for _, item := range result.Items {
		var laborLine models.LaborLine
		err = attributevalue.UnmarshalMap(item, &laborLine)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling labor line: %w", err)
		}

		// Skip soft-deleted items
		if !laborLine.IsDeleted() {
			laborLines = append(laborLines, &laborLine)
		}
	}

	return laborLines, nil
}
