package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"steverhoton-labor-lines/lambda/models"
)

// MockDynamoDBService is a mock implementation of DynamoDBService.
type MockDynamoDBService struct {
	mock.Mock
}

func (m *MockDynamoDBService) CreateLaborLine(ctx context.Context, laborLine *models.LaborLine) error {
	args := m.Called(ctx, laborLine)
	return args.Error(0)
}

func (m *MockDynamoDBService) GetLaborLine(ctx context.Context, input models.GetLaborLineInput) (*models.LaborLine, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*models.LaborLine), args.Error(1)
}

func (m *MockDynamoDBService) UpdateLaborLine(ctx context.Context, laborLine *models.LaborLine) error {
	args := m.Called(ctx, laborLine)
	return args.Error(0)
}

func (m *MockDynamoDBService) DeleteLaborLine(ctx context.Context, input models.DeleteLaborLineInput) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

func (m *MockDynamoDBService) ListLaborLines(ctx context.Context, input models.ListLaborLinesInput) ([]*models.LaborLine, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]*models.LaborLine), args.Error(1)
}

// MockValidationService is a mock implementation of ValidationService.
type MockValidationService struct {
	mock.Mock
}

func (m *MockValidationService) ValidateCreateInput(input models.CreateLaborLineInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func (m *MockValidationService) ValidateUpdateInput(input models.UpdateLaborLineInput) error {
	args := m.Called(input)
	return args.Error(0)
}

func TestNewLaborLineHandler(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}

	handler := NewLaborLineHandler(dynamoDBService, validationService)
	assert.NotNil(t, handler)
}

func TestLaborLineHandler_HandleAppSyncEvent_CreateLaborLine(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	input := models.CreateLaborLineInput{
		ContactID: uuid.New().String(),
		AccountID: uuid.New().String(),
		TaskID:    uuid.New().String(),
	}

	event := models.AppSyncEvent{
		Info: models.AppSyncEventInfo{
			FieldName:      "createLaborLine",
			ParentTypeName: "Mutation",
		},
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"contactId": input.ContactID,
				"accountId": input.AccountID,
				"taskId":    input.TaskID,
			},
		},
	}

	validationService.On("ValidateCreateInput", mock.MatchedBy(func(i models.CreateLaborLineInput) bool {
		return i.ContactID == input.ContactID && i.AccountID == input.AccountID && i.TaskID == input.TaskID
	})).Return(nil)

	dynamoDBService.On("CreateLaborLine", mock.Anything, mock.MatchedBy(func(ll *models.LaborLine) bool {
		return ll.ContactID == input.ContactID && ll.AccountID == input.AccountID && ll.TaskID == input.TaskID
	})).Return(nil)

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Data)

	// Verify the response contains a labor line
	laborLine, ok := response.Data.(*models.LaborLine)
	require.True(t, ok)
	assert.Equal(t, input.ContactID, laborLine.ContactID)
	assert.Equal(t, input.AccountID, laborLine.AccountID)
	assert.Equal(t, input.TaskID, laborLine.TaskID)

	dynamoDBService.AssertExpectations(t)
	validationService.AssertExpectations(t)
}

func TestLaborLineHandler_HandleAppSyncEvent_CreateLaborLine_ValidationError(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	event := models.AppSyncEvent{
		FieldName: "createLaborLine",
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"contactId": "invalid-uuid",
				"accountId": uuid.New().String(),
				"taskId":    uuid.New().String(),
			},
		},
	}

	validationService.On("ValidateCreateInput", mock.Anything).Return(fmt.Errorf("validation failed"))

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Error)
	assert.Equal(t, "ValidationError", response.Error.Type)
	assert.Contains(t, response.Error.Message, "validation failed")

	validationService.AssertExpectations(t)
}

func TestLaborLineHandler_HandleAppSyncEvent_GetLaborLine(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	accountID := uuid.New().String()
	taskID := uuid.New().String()
	laborLineID := uuid.New().String()

	expectedLaborLine := &models.LaborLine{
		LaborLineID: laborLineID,
		ContactID:   uuid.New().String(),
		AccountID:   accountID,
		TaskID:      taskID,
	}

	event := models.AppSyncEvent{
		FieldName: "getLaborLine",
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"accountId":   accountID,
				"taskId":      taskID,
				"laborLineId": laborLineID,
			},
		},
	}

	dynamoDBService.On("GetLaborLine", mock.Anything, models.GetLaborLineInput{
		AccountID:   accountID,
		TaskID:      taskID,
		LaborLineID: laborLineID,
	}).Return(expectedLaborLine, nil)

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Nil(t, response.Error)
	assert.Equal(t, expectedLaborLine, response.Data)

	dynamoDBService.AssertExpectations(t)
}

func TestLaborLineHandler_HandleAppSyncEvent_GetLaborLine_NotFound(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	event := models.AppSyncEvent{
		FieldName: "getLaborLine",
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"accountId":   uuid.New().String(),
				"taskId":      uuid.New().String(),
				"laborLineId": uuid.New().String(),
			},
		},
	}

	dynamoDBService.On("GetLaborLine", mock.Anything, mock.Anything).Return((*models.LaborLine)(nil), nil)

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Error)
	assert.Equal(t, "NotFound", response.Error.Type)
	assert.Equal(t, "labor line not found", response.Error.Message)

	dynamoDBService.AssertExpectations(t)
}

func TestLaborLineHandler_HandleAppSyncEvent_UpdateLaborLine(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	input := models.UpdateLaborLineInput{
		LaborLineID: uuid.New().String(),
		ContactID:   uuid.New().String(),
		AccountID:   uuid.New().String(),
		TaskID:      uuid.New().String(),
	}

	updatedLaborLine := &models.LaborLine{
		LaborLineID: input.LaborLineID,
		ContactID:   input.ContactID,
		AccountID:   input.AccountID,
		TaskID:      input.TaskID,
	}

	event := models.AppSyncEvent{
		FieldName: "updateLaborLine",
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"laborLineId": input.LaborLineID,
				"contactId":   input.ContactID,
				"accountId":   input.AccountID,
				"taskId":      input.TaskID,
			},
		},
	}

	validationService.On("ValidateUpdateInput", mock.Anything).Return(nil)
	dynamoDBService.On("UpdateLaborLine", mock.Anything, mock.Anything).Return(nil)
	dynamoDBService.On("GetLaborLine", mock.Anything, mock.Anything).Return(updatedLaborLine, nil)

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Nil(t, response.Error)
	assert.Equal(t, updatedLaborLine, response.Data)

	dynamoDBService.AssertExpectations(t)
	validationService.AssertExpectations(t)
}

func TestLaborLineHandler_HandleAppSyncEvent_DeleteLaborLine(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	event := models.AppSyncEvent{
		FieldName: "deleteLaborLine",
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"accountId":   uuid.New().String(),
				"taskId":      uuid.New().String(),
				"laborLineId": uuid.New().String(),
			},
		},
	}

	dynamoDBService.On("DeleteLaborLine", mock.Anything, mock.Anything).Return(nil)

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Nil(t, response.Error)

	// Verify success response
	data, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, data["success"])
	assert.Equal(t, "labor line deleted successfully", data["message"])

	dynamoDBService.AssertExpectations(t)
}

func TestLaborLineHandler_HandleAppSyncEvent_ListLaborLines(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	accountID := uuid.New().String()
	expectedLaborLines := []*models.LaborLine{
		{
			LaborLineID: uuid.New().String(),
			ContactID:   uuid.New().String(),
			AccountID:   accountID,
			TaskID:      uuid.New().String(),
		},
		{
			LaborLineID: uuid.New().String(),
			ContactID:   uuid.New().String(),
			AccountID:   accountID,
			TaskID:      uuid.New().String(),
		},
	}

	event := models.AppSyncEvent{
		FieldName: "listLaborLines",
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"accountId": accountID,
			},
		},
	}

	dynamoDBService.On("ListLaborLines", mock.Anything, mock.Anything).Return(expectedLaborLines, nil)

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Nil(t, response.Error)
	assert.Equal(t, expectedLaborLines, response.Data)

	dynamoDBService.AssertExpectations(t)
}

func TestLaborLineHandler_HandleAppSyncEvent_UnsupportedOperation(t *testing.T) {
	dynamoDBService := &MockDynamoDBService{}
	validationService := &MockValidationService{}
	handler := NewLaborLineHandler(dynamoDBService, validationService)

	event := models.AppSyncEvent{
		FieldName: "unsupportedOperation",
		Arguments: map[string]interface{}{},
	}

	response, err := handler.HandleAppSyncEvent(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Error)
	assert.Equal(t, "UnsupportedOperation", response.Error.Type)
	assert.Contains(t, response.Error.Message, "unsupportedOperation")
}
