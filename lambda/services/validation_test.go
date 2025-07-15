package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"steverhoton-labor-lines/lambda/models"
)

func TestNewValidationServiceWithEmbeddedSchema(t *testing.T) {
	validationService := NewValidationServiceWithEmbeddedSchema()
	assert.NotNil(t, validationService)
}

func TestValidationService_ValidateCreateInput(t *testing.T) {
	validationService := NewValidationServiceWithEmbeddedSchema()

	tests := []struct {
		name      string
		input     models.CreateLaborLineInput
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid input with all fields",
			input: models.CreateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    uuid.New().String(),
				PartID:    []string{uuid.New().String(), uuid.New().String()},
				Notes:     []string{"Valid note", "Another valid note"},
			},
			wantError: false,
		},
		{
			name: "Valid input with required fields only",
			input: models.CreateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    uuid.New().String(),
			},
			wantError: false,
		},
		{
			name: "Missing accountId",
			input: models.CreateLaborLineInput{
				TaskID: uuid.New().String(),
			},
			wantError: true,
		},
		{
			name: "Missing taskId",
			input: models.CreateLaborLineInput{
				AccountID: uuid.New().String(),
			},
			wantError: true,
		},
		{
			name: "Invalid accountId UUID",
			input: models.CreateLaborLineInput{
				AccountID: "invalid-uuid",
				TaskID:    uuid.New().String(),
			},
			wantError: true,
			errorMsg:  "invalid UUID format",
		},
		{
			name: "Invalid taskId UUID",
			input: models.CreateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    "invalid-uuid",
			},
			wantError: true,
			errorMsg:  "invalid UUID format",
		},
		{
			name: "Invalid partId UUID",
			input: models.CreateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    uuid.New().String(),
				PartID:    []string{"invalid-uuid"},
			},
			wantError: true,
			errorMsg:  "invalid UUID format",
		},
		{
			name: "Empty note",
			input: models.CreateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    uuid.New().String(),
				Notes:     []string{""},
			},
			wantError: true,
		},
		{
			name: "Note too long",
			input: models.CreateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    uuid.New().String(),
				Notes:     []string{generateLongString(1001)},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationService.ValidateCreateInput(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationService_ValidateUpdateInput(t *testing.T) {
	validationService := NewValidationServiceWithEmbeddedSchema()

	tests := []struct {
		name      string
		input     models.UpdateLaborLineInput
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid update input",
			input: models.UpdateLaborLineInput{
				LaborLineID: uuid.New().String(),
				AccountID:   uuid.New().String(),
				TaskID:      uuid.New().String(),
				PartID:      []string{uuid.New().String()},
				Notes:       []string{"Updated note"},
			},
			wantError: false,
		},
		{
			name: "Valid update input with required fields only",
			input: models.UpdateLaborLineInput{
				LaborLineID: uuid.New().String(),
				AccountID:   uuid.New().String(),
				TaskID:      uuid.New().String(),
			},
			wantError: false,
		},
		{
			name: "Missing laborLineId",
			input: models.UpdateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    uuid.New().String(),
			},
			wantError: true,
		},
		{
			name: "Invalid laborLineId UUID",
			input: models.UpdateLaborLineInput{
				LaborLineID: "invalid-uuid",
				AccountID:   uuid.New().String(),
				TaskID:      uuid.New().String(),
			},
			wantError: true,
			errorMsg:  "invalid UUID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationService.ValidateUpdateInput(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationService_validateUUIDs(t *testing.T) {
	validationService := NewValidationServiceWithEmbeddedSchema().(*validationService)

	tests := []struct {
		name      string
		data      map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid UUIDs",
			data: map[string]interface{}{
				"laborLineId": uuid.New().String(),
				"accountId":   uuid.New().String(),
				"taskId":      uuid.New().String(),
				"partId":      []string{uuid.New().String(), uuid.New().String()},
			},
			wantError: false,
		},
		{
			name: "Invalid laborLineId",
			data: map[string]interface{}{
				"laborLineId": "invalid-uuid",
			},
			wantError: true,
			errorMsg:  "invalid UUID format for field laborLineId",
		},
		{
			name: "Invalid partId UUID",
			data: map[string]interface{}{
				"partId": []string{"invalid-uuid"},
			},
			wantError: true,
			errorMsg:  "invalid UUID format for partId[0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validationService.validateUUIDs(tt.data)

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// generateLongString creates a string of specified length for testing.
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}
