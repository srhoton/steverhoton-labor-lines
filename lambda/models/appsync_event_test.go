package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSyncEvent_GetArgumentAs(t *testing.T) {
	tests := []struct {
		name      string
		event     AppSyncEvent
		key       string
		target    interface{}
		expected  interface{}
		wantError bool
	}{
		{
			name: "String argument",
			event: AppSyncEvent{
				Arguments: map[string]interface{}{
					"testKey": "testValue",
				},
			},
			key:      "testKey",
			target:   new(string),
			expected: "testValue",
		},
		{
			name: "Map argument",
			event: AppSyncEvent{
				Arguments: map[string]interface{}{
					"input": map[string]interface{}{
						"accountId": uuid.New().String(),
						"taskId":    uuid.New().String(),
					},
				},
			},
			key:    "input",
			target: new(map[string]interface{}),
		},
		{
			name: "Non-existent key",
			event: AppSyncEvent{
				Arguments: map[string]interface{}{},
			},
			key:      "nonExistent",
			target:   new(string),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.GetArgumentAs(tt.key, tt.target)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.expected != nil {
				switch target := tt.target.(type) {
				case *string:
					assert.Equal(t, tt.expected, *target)
				case *map[string]interface{}:
					assert.NotNil(t, *target)
				}
			}
		})
	}
}

func TestAppSyncEvent_GetInputArgument(t *testing.T) {
	accountID := uuid.New().String()
	taskID := uuid.New().String()

	event := AppSyncEvent{
		Arguments: map[string]interface{}{
			"input": map[string]interface{}{
				"accountId": accountID,
				"taskId":    taskID,
				"notes":     []interface{}{"Test note"},
			},
		},
	}

	var input CreateLaborLineInput
	err := event.GetInputArgument(&input)

	require.NoError(t, err)
	assert.Equal(t, accountID, input.AccountID)
	assert.Equal(t, taskID, input.TaskID)
	assert.Equal(t, []string{"Test note"}, input.Notes)
}

func TestAppSyncEvent_GetInputArgument_NoInput(t *testing.T) {
	event := AppSyncEvent{
		Arguments: map[string]interface{}{},
	}

	var input CreateLaborLineInput
	err := event.GetInputArgument(&input)

	assert.NoError(t, err)
	// Should be zero value since no input was provided
	assert.Empty(t, input.AccountID)
	assert.Empty(t, input.TaskID)
}
