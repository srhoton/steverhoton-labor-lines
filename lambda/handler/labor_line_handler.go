// Package handler contains the Lambda function handlers.
package handler

import (
	"context"
	"fmt"
	"log"

	"steverhoton-labor-lines/lambda/models"
	"steverhoton-labor-lines/lambda/services"
)

// LaborLineHandler handles AppSync events for labor line operations.
type LaborLineHandler struct {
	dynamoDBService   services.DynamoDBService
	validationService services.ValidationService
}

// NewLaborLineHandler creates a new labor line handler.
func NewLaborLineHandler(dynamoDBService services.DynamoDBService, validationService services.ValidationService) *LaborLineHandler {
	return &LaborLineHandler{
		dynamoDBService:   dynamoDBService,
		validationService: validationService,
	}
}

// HandleAppSyncEvent processes AppSync events and routes them to appropriate handlers.
func (h *LaborLineHandler) HandleAppSyncEvent(ctx context.Context, event models.AppSyncEvent) (*models.AppSyncResponse, error) {
	// AppSync Direct Lambda Resolvers send field information in the info object
	fieldName := event.Info.FieldName
	typeName := event.Info.ParentTypeName

	log.Printf("Processing AppSync event: %s.%s", typeName, fieldName)

	switch fieldName {
	case "createLaborLine":
		return h.handleCreate(ctx, event)
	case "updateLaborLine":
		return h.handleUpdate(ctx, event)
	case "deleteLaborLine":
		return h.handleDelete(ctx, event)
	case "getLaborLine":
		return h.handleGet(ctx, event)
	case "listLaborLines":
		return h.handleList(ctx, event)
	default:
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("unsupported operation: %s", fieldName),
				Type:    "UnsupportedOperation",
			},
		}, nil
	}
}

// handleCreate processes create labor line requests.
func (h *LaborLineHandler) handleCreate(ctx context.Context, event models.AppSyncEvent) (*models.AppSyncResponse, error) {
	var input models.CreateLaborLineInput
	if err := event.GetInputArgument(&input); err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("invalid input: %v", err),
				Type:    "ValidationError",
			},
		}, nil
	}

	// Validate input
	if err := h.validationService.ValidateCreateInput(input); err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("validation failed: %v", err),
				Type:    "ValidationError",
			},
		}, nil
	}

	// Create labor line
	laborLine := models.NewLaborLine(input)
	if err := h.dynamoDBService.CreateLaborLine(ctx, laborLine); err != nil {
		log.Printf("Error creating labor line: %v", err)
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "failed to create labor line",
				Type:    "InternalError",
			},
		}, nil
	}

	return &models.AppSyncResponse{
		Data: laborLine,
	}, nil
}

// handleUpdate processes update labor line requests.
func (h *LaborLineHandler) handleUpdate(ctx context.Context, event models.AppSyncEvent) (*models.AppSyncResponse, error) {
	var input models.UpdateLaborLineInput
	if err := event.GetInputArgument(&input); err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("invalid input: %v", err),
				Type:    "ValidationError",
			},
		}, nil
	}

	// Validate input
	if err := h.validationService.ValidateUpdateInput(input); err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("validation failed: %v", err),
				Type:    "ValidationError",
			},
		}, nil
	}

	// Update labor line
	laborLine := input.ToLaborLine()
	if err := h.dynamoDBService.UpdateLaborLine(ctx, laborLine); err != nil {
		log.Printf("Error updating labor line: %v", err)
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "failed to update labor line",
				Type:    "InternalError",
			},
		}, nil
	}

	// Return the updated labor line
	updatedLaborLine, err := h.dynamoDBService.GetLaborLine(ctx, models.GetLaborLineInput{
		AccountID:   input.AccountID,
		TaskID:      input.TaskID,
		LaborLineID: input.LaborLineID,
	})
	if err != nil {
		log.Printf("Error retrieving updated labor line: %v", err)
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "failed to retrieve updated labor line",
				Type:    "InternalError",
			},
		}, nil
	}

	return &models.AppSyncResponse{
		Data: updatedLaborLine,
	}, nil
}

// handleDelete processes delete labor line requests.
func (h *LaborLineHandler) handleDelete(ctx context.Context, event models.AppSyncEvent) (*models.AppSyncResponse, error) {
	var input models.DeleteLaborLineInput
	if err := event.GetInputArgument(&input); err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("invalid input: %v", err),
				Type:    "ValidationError",
			},
		}, nil
	}

	// Delete labor line
	if err := h.dynamoDBService.DeleteLaborLine(ctx, input); err != nil {
		log.Printf("Error deleting labor line: %v", err)
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "failed to delete labor line",
				Type:    "InternalError",
			},
		}, nil
	}

	return &models.AppSyncResponse{
		Data: map[string]interface{}{
			"success": true,
			"message": "labor line deleted successfully",
		},
	}, nil
}

// handleGet processes get labor line requests.
func (h *LaborLineHandler) handleGet(ctx context.Context, event models.AppSyncEvent) (*models.AppSyncResponse, error) {
	var input models.GetLaborLineInput
	if err := event.GetInputArgument(&input); err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("invalid input: %v", err),
				Type:    "ValidationError",
			},
		}, nil
	}

	// Get labor line
	laborLine, err := h.dynamoDBService.GetLaborLine(ctx, input)
	if err != nil {
		log.Printf("Error getting labor line: %v", err)
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "failed to get labor line",
				Type:    "InternalError",
			},
		}, nil
	}

	if laborLine == nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "labor line not found",
				Type:    "NotFound",
			},
		}, nil
	}

	return &models.AppSyncResponse{
		Data: laborLine,
	}, nil
}

// handleList processes list labor lines requests.
func (h *LaborLineHandler) handleList(ctx context.Context, event models.AppSyncEvent) (*models.AppSyncResponse, error) {
	var input models.ListLaborLinesInput
	if err := event.GetInputArgument(&input); err != nil {
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: fmt.Sprintf("invalid input: %v", err),
				Type:    "ValidationError",
			},
		}, nil
	}

	// List labor lines
	laborLines, err := h.dynamoDBService.ListLaborLines(ctx, input)
	if err != nil {
		log.Printf("Error listing labor lines: %v", err)
		return &models.AppSyncResponse{
			Error: &models.AppSyncError{
				Message: "failed to list labor lines",
				Type:    "InternalError",
			},
		}, nil
	}

	return &models.AppSyncResponse{
		Data: laborLines,
	}, nil
}
