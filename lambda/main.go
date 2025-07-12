// Package main contains the entry point for the Lambda function.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"steverhoton-labor-lines/lambda/handler"
	"steverhoton-labor-lines/lambda/models"
	"steverhoton-labor-lines/lambda/services"
)

// LambdaHandler is the main Lambda function handler.
func LambdaHandler(ctx context.Context, event models.AppSyncEvent) (*models.AppSyncResponse, error) {
	// Get table name from environment variable
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "DYNAMODB_TABLE_NAME environment variable not set",
				Type:    "ConfigurationError",
			},
		}, nil
	}

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("failed to load AWS config: %v", err),
				Type:    "ConfigurationError",
			},
		}, nil
	}

	// Create DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Create services
	dynamoDBService := services.NewDynamoDBService(dynamoClient, tableName)
	validationService := services.NewValidationServiceWithEmbeddedSchema()

	// Create handler
	laborLineHandler := handler.NewLaborLineHandler(dynamoDBService, validationService)

	// Process the event
	return laborLineHandler.HandleAppSyncEvent(ctx, event)
}

func main() {
	lambda.Start(LambdaHandler)
}
