package services

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"steverhoton-labor-lines/lambda/models"
)

// MockDynamoDBClient is a mock implementation of DynamoDBClient.
type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func TestNewDynamoDBService(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"

	service := NewDynamoDBService(client, tableName)
	assert.NotNil(t, service)
}

func TestDynamoDBService_CreateLaborLine(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"
	service := NewDynamoDBService(client, tableName)

	laborLine := &models.LaborLine{
		LaborLineID: uuid.New().String(),
		AccountID:   uuid.New().String(),
		TaskID:      uuid.New().String(),
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		PK:          uuid.New().String(),
		SK:          uuid.New().String() + "#" + uuid.New().String(),
	}

	client.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		return *input.TableName == tableName && input.ConditionExpression != nil
	})).Return(&dynamodb.PutItemOutput{}, nil)

	err := service.CreateLaborLine(context.Background(), laborLine)
	assert.NoError(t, err)

	client.AssertExpectations(t)
}

func TestDynamoDBService_GetLaborLine(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"
	service := NewDynamoDBService(client, tableName)

	accountID := uuid.New().String()
	taskID := uuid.New().String()
	laborLineID := uuid.New().String()

	laborLine := &models.LaborLine{
		LaborLineID: laborLineID,
		AccountID:   accountID,
		TaskID:      taskID,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		PK:          accountID,
		SK:          taskID + "#" + laborLineID,
	}

	item, _ := attributevalue.MarshalMap(laborLine)

	client.On("GetItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.GetItemInput) bool {
		return *input.TableName == tableName
	})).Return(&dynamodb.GetItemOutput{Item: item}, nil)

	input := models.GetLaborLineInput{
		AccountID:   accountID,
		TaskID:      taskID,
		LaborLineID: laborLineID,
	}

	result, err := service.GetLaborLine(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, laborLineID, result.LaborLineID)
	assert.Equal(t, accountID, result.AccountID)
	assert.Equal(t, taskID, result.TaskID)

	client.AssertExpectations(t)
}

func TestDynamoDBService_GetLaborLine_NotFound(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"
	service := NewDynamoDBService(client, tableName)

	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)

	input := models.GetLaborLineInput{
		AccountID:   uuid.New().String(),
		TaskID:      uuid.New().String(),
		LaborLineID: uuid.New().String(),
	}

	result, err := service.GetLaborLine(context.Background(), input)
	assert.NoError(t, err)
	assert.Nil(t, result)

	client.AssertExpectations(t)
}

func TestDynamoDBService_GetLaborLine_SoftDeleted(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"
	service := NewDynamoDBService(client, tableName)

	accountID := uuid.New().String()
	taskID := uuid.New().String()
	laborLineID := uuid.New().String()
	deletedAt := time.Now().Unix()

	laborLine := &models.LaborLine{
		LaborLineID: laborLineID,
		AccountID:   accountID,
		TaskID:      taskID,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		DeletedAt:   &deletedAt,
		PK:          accountID,
		SK:          taskID + "#" + laborLineID,
	}

	item, _ := attributevalue.MarshalMap(laborLine)

	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: item}, nil)

	input := models.GetLaborLineInput{
		AccountID:   accountID,
		TaskID:      taskID,
		LaborLineID: laborLineID,
	}

	result, err := service.GetLaborLine(context.Background(), input)
	assert.NoError(t, err)
	assert.Nil(t, result) // Should return nil for soft-deleted items

	client.AssertExpectations(t)
}

func TestDynamoDBService_UpdateLaborLine(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"
	service := NewDynamoDBService(client, tableName)

	accountID := uuid.New().String()
	taskID := uuid.New().String()
	laborLineID := uuid.New().String()

	existingLaborLine := &models.LaborLine{
		LaborLineID: laborLineID,
		AccountID:   accountID,
		TaskID:      taskID,
		CreatedAt:   time.Now().Unix() - 100,
		UpdatedAt:   time.Now().Unix() - 50,
		PK:          accountID,
		SK:          taskID + "#" + laborLineID,
	}

	updateLaborLine := &models.LaborLine{
		LaborLineID: laborLineID,
		AccountID:   accountID,
		TaskID:      taskID,
		UpdatedAt:   time.Now().Unix(),
		PK:          accountID,
		SK:          taskID + "#" + laborLineID,
	}

	existingItem, _ := attributevalue.MarshalMap(existingLaborLine)

	// Mock GetItem call for checking existing item
	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: existingItem}, nil)

	// Mock PutItem call for update
	client.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		return *input.TableName == tableName && input.ConditionExpression != nil
	})).Return(&dynamodb.PutItemOutput{}, nil)

	err := service.UpdateLaborLine(context.Background(), updateLaborLine)
	assert.NoError(t, err)

	client.AssertExpectations(t)
}

func TestDynamoDBService_DeleteLaborLine(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"
	service := NewDynamoDBService(client, tableName)

	accountID := uuid.New().String()
	taskID := uuid.New().String()
	laborLineID := uuid.New().String()

	existingLaborLine := &models.LaborLine{
		LaborLineID: laborLineID,
		AccountID:   accountID,
		TaskID:      taskID,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		PK:          accountID,
		SK:          taskID + "#" + laborLineID,
	}

	existingItem, _ := attributevalue.MarshalMap(existingLaborLine)

	// Mock GetItem call for checking existing item
	client.On("GetItem", mock.Anything, mock.Anything).Return(&dynamodb.GetItemOutput{Item: existingItem}, nil)

	// Mock PutItem call for soft delete
	client.On("PutItem", mock.Anything, mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
		return *input.TableName == tableName && input.ConditionExpression != nil
	})).Return(&dynamodb.PutItemOutput{}, nil)

	input := models.DeleteLaborLineInput{
		AccountID:   accountID,
		TaskID:      taskID,
		LaborLineID: laborLineID,
	}

	err := service.DeleteLaborLine(context.Background(), input)
	assert.NoError(t, err)

	client.AssertExpectations(t)
}

func TestDynamoDBService_ListLaborLines(t *testing.T) {
	client := &MockDynamoDBClient{}
	tableName := "test-table"
	service := NewDynamoDBService(client, tableName)

	accountID := uuid.New().String()
	taskID := uuid.New().String()

	laborLine1 := &models.LaborLine{
		LaborLineID: uuid.New().String(),
		AccountID:   accountID,
		TaskID:      taskID,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		PK:          accountID,
		SK:          taskID + "#" + uuid.New().String(),
	}

	laborLine2 := &models.LaborLine{
		LaborLineID: uuid.New().String(),
		AccountID:   accountID,
		TaskID:      taskID,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
		PK:          accountID,
		SK:          taskID + "#" + uuid.New().String(),
	}

	item1, _ := attributevalue.MarshalMap(laborLine1)
	item2, _ := attributevalue.MarshalMap(laborLine2)

	client.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
		return *input.TableName == tableName
	})).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{item1, item2},
	}, nil)

	tests := []struct {
		name  string
		input models.ListLaborLinesInput
	}{
		{
			name: "List by account only",
			input: models.ListLaborLinesInput{
				AccountID: accountID,
			},
		},
		{
			name: "List by account and task",
			input: models.ListLaborLinesInput{
				AccountID: accountID,
				TaskID:    taskID,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ListLaborLines(context.Background(), tt.input)
			require.NoError(t, err)
			assert.Len(t, result, 2)
		})
	}

	client.AssertExpectations(t)
}
