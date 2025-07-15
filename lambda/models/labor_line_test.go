package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLaborLine(t *testing.T) {
	tests := []struct {
		name  string
		input CreateLaborLineInput
	}{
		{
			name: "Valid input with all fields",
			input: CreateLaborLineInput{
				AccountID:   uuid.New().String(),
				TaskID:      uuid.New().String(),
				PartID:      []string{uuid.New().String(), uuid.New().String()},
				Notes:       []string{"First note", "Second note"},
				Description: "Complete brake system maintenance",
			},
		},
		{
			name: "Valid input with required fields only",
			input: CreateLaborLineInput{
				AccountID: uuid.New().String(),
				TaskID:    uuid.New().String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startTime := time.Now().Unix()
			laborLine := NewLaborLine(tt.input)

			// Verify required fields
			assert.NotEmpty(t, laborLine.LaborLineID)
			assert.Equal(t, tt.input.AccountID, laborLine.AccountID)
			assert.Equal(t, tt.input.TaskID, laborLine.TaskID)

			// Verify optional fields
			assert.Equal(t, tt.input.PartID, laborLine.PartID)
			assert.Equal(t, tt.input.Notes, laborLine.Notes)
			assert.Equal(t, tt.input.Description, laborLine.Description)

			// Verify timestamps
			assert.GreaterOrEqual(t, laborLine.CreatedAt, startTime)
			assert.GreaterOrEqual(t, laborLine.UpdatedAt, startTime)
			assert.Equal(t, laborLine.CreatedAt, laborLine.UpdatedAt)
			assert.Nil(t, laborLine.DeletedAt)

			// Verify DynamoDB keys
			assert.Equal(t, tt.input.AccountID, laborLine.PK)
			assert.Equal(t, tt.input.TaskID+"#"+laborLine.LaborLineID, laborLine.SK)

			// Verify LaborLineID is a valid UUID
			_, err := uuid.Parse(laborLine.LaborLineID)
			require.NoError(t, err)
		})
	}
}

func TestUpdateLaborLineInput_ToLaborLine(t *testing.T) {
	input := UpdateLaborLineInput{
		LaborLineID: uuid.New().String(),
		AccountID:   uuid.New().String(),
		TaskID:      uuid.New().String(),
		PartID:      []string{uuid.New().String()},
		Notes:       []string{"Updated note"},
		Description: "Updated brake system maintenance task",
	}

	startTime := time.Now().Unix()
	laborLine := input.ToLaborLine()

	// Verify all fields are set correctly
	assert.Equal(t, input.LaborLineID, laborLine.LaborLineID)
	assert.Equal(t, input.AccountID, laborLine.AccountID)
	assert.Equal(t, input.TaskID, laborLine.TaskID)
	assert.Equal(t, input.PartID, laborLine.PartID)
	assert.Equal(t, input.Notes, laborLine.Notes)
	assert.Equal(t, input.Description, laborLine.Description)

	// Verify DynamoDB keys
	assert.Equal(t, input.AccountID, laborLine.PK)
	assert.Equal(t, input.TaskID+"#"+input.LaborLineID, laborLine.SK)

	// Verify UpdatedAt is set
	assert.GreaterOrEqual(t, laborLine.UpdatedAt, startTime)

	// Verify CreatedAt and DeletedAt are not set (will be set during update)
	assert.Zero(t, laborLine.CreatedAt)
	assert.Nil(t, laborLine.DeletedAt)
}

func TestLaborLine_IsDeleted(t *testing.T) {
	tests := []struct {
		name      string
		deletedAt *int64
		expected  bool
	}{
		{
			name:      "Not deleted",
			deletedAt: nil,
			expected:  false,
		},
		{
			name:      "Deleted",
			deletedAt: func() *int64 { t := time.Now().Unix(); return &t }(),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			laborLine := &LaborLine{
				DeletedAt: tt.deletedAt,
			}

			assert.Equal(t, tt.expected, laborLine.IsDeleted())
		})
	}
}

func TestLaborLine_SoftDelete(t *testing.T) {
	laborLine := &LaborLine{
		LaborLineID: uuid.New().String(),
		CreatedAt:   time.Now().Unix() - 100,
		UpdatedAt:   time.Now().Unix() - 50,
	}

	startTime := time.Now().Unix()
	laborLine.SoftDelete()

	// Verify DeletedAt is set
	require.NotNil(t, laborLine.DeletedAt)
	assert.GreaterOrEqual(t, *laborLine.DeletedAt, startTime)

	// Verify UpdatedAt is updated
	assert.GreaterOrEqual(t, laborLine.UpdatedAt, startTime)

	// Verify IsDeleted returns true
	assert.True(t, laborLine.IsDeleted())
}
