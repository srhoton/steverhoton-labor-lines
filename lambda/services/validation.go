package services

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/xeipuuv/gojsonschema"

	"steverhoton-labor-lines/lambda/models"
)

// ValidationService defines the interface for validation operations.
type ValidationService interface {
	ValidateCreateInput(input models.CreateLaborLineInput) error
	ValidateUpdateInput(input models.UpdateLaborLineInput) error
}

// validationService implements ValidationService.
type validationService struct {
	schema *gojsonschema.Schema
}

// NewValidationService creates a new validation service instance.
func NewValidationService(schemaPath string) (ValidationService, error) {
	// Read schema file
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("reading schema file: %w", err)
	}

	// Load schema
	schemaLoader := gojsonschema.NewBytesLoader(schemaBytes)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("loading JSON schema: %w", err)
	}

	return &validationService{
		schema: schema,
	}, nil
}

// NewValidationServiceWithEmbeddedSchema creates a validation service with embedded schema.
// This is useful for deployment where we don't want to read files at runtime.
func NewValidationServiceWithEmbeddedSchema() ValidationService {
	// Embedded schema JSON - matches the labor-line.schema.json file
	schemaJSON := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"$id": "https://example.com/schemas/labor-line.schema.json",
		"title": "Labor Line",
		"description": "A labor line for maintenance work order tasks",
		"type": "object",
		"properties": {
			"laborLineId": {
				"type": "string",
				"format": "uuid",
				"description": "Unique identifier for the labor line"
			},
			"accountId": {
				"type": "string",
				"format": "uuid",
				"description": "Account identifier (used as DynamoDB partition key)"
			},
			"taskId": {
				"type": "string",
				"format": "uuid",
				"description": "Task identifier (used in DynamoDB sort key)"
			},
			"partId": {
				"type": "array",
				"items": {
					"type": "string",
					"format": "uuid"
				},
				"description": "Optional list of part identifiers required for the work",
				"uniqueItems": true
			},
			"notes": {
				"type": "array",
				"items": {
					"type": "string",
					"minLength": 1,
					"maxLength": 1000
				},
				"description": "Optional notes describing the work to be performed"
			}
		},
		"required": [
			"laborLineId",
			"accountId",
			"taskId"
		],
		"additionalProperties": false
	}`

	schemaLoader := gojsonschema.NewStringLoader(schemaJSON)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("loading embedded JSON schema: %w", err)
	}

	return &validationService{
		schema: schema,
	}, nil
}

// ValidateCreateInput validates a CreateLaborLineInput against the JSON schema.
func (s *validationService) ValidateCreateInput(input models.CreateLaborLineInput) error {
	// Convert to a map that includes a generated laborLineId for validation
	validationData := map[string]interface{}{
		"laborLineId": uuid.New().String(), // Temporary ID for validation
		"accountId":   input.AccountID,
		"taskId":      input.TaskID,
	}

	if input.PartID != nil {
		validationData["partId"] = input.PartID
	}
	if input.Notes != nil {
		validationData["notes"] = input.Notes
	}

	return s.validateData(validationData)
}

// ValidateUpdateInput validates an UpdateLaborLineInput against the JSON schema.
func (s *validationService) ValidateUpdateInput(input models.UpdateLaborLineInput) error {
	validationData := map[string]interface{}{
		"laborLineId": input.LaborLineID,
		"accountId":   input.AccountID,
		"taskId":      input.TaskID,
	}

	if input.PartID != nil {
		validationData["partId"] = input.PartID
	}
	if input.Notes != nil {
		validationData["notes"] = input.Notes
	}

	return s.validateData(validationData)
}

// validateData validates the given data against the JSON schema.
func (s *validationService) validateData(data map[string]interface{}) error {
	// Additional UUID validation
	if err := s.validateUUIDs(data); err != nil {
		return err
	}

	// Validate against JSON schema
	dataLoader := gojsonschema.NewGoLoader(data)
	result, err := s.schema.Validate(dataLoader)
	if err != nil {
		return fmt.Errorf("schema validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, desc.String())
		}
		return fmt.Errorf("validation failed: %v", errors)
	}

	return nil
}

// validateUUIDs validates that all UUID fields are properly formatted.
func (s *validationService) validateUUIDs(data map[string]interface{}) error {
	uuidFields := []string{"laborLineId", "accountId", "taskId"}

	for _, field := range uuidFields {
		if value, exists := data[field]; exists {
			if strValue, ok := value.(string); ok {
				if _, err := uuid.Parse(strValue); err != nil {
					return fmt.Errorf("invalid UUID format for field %s: %s", field, strValue)
				}
			}
		}
	}

	// Validate partId array if present
	if partID, exists := data["partId"]; exists {
		if partArray, ok := partID.([]string); ok {
			for i, part := range partArray {
				if _, err := uuid.Parse(part); err != nil {
					return fmt.Errorf("invalid UUID format for partId[%d]: %s", i, part)
				}
			}
		}
	}

	return nil
}
