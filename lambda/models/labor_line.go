// Package models contains data structures for the labor lines service.
package models

import (
	"time"

	"github.com/google/uuid"
)

// LaborLine represents a maintenance labor line for work order tasks.
// It includes all fields from the JSON schema plus audit timestamps.
type LaborLine struct {
	// Required fields from schema
	LaborLineID string `json:"laborLineId" dynamodbav:"laborLineId"`
	ContactID   string `json:"contactId" dynamodbav:"contactId"`
	AccountID   string `json:"accountId" dynamodbav:"accountId"`
	TaskID      string `json:"taskId" dynamodbav:"taskId"`

	// Optional fields from schema
	PartID []string `json:"partId,omitempty" dynamodbav:"partId,omitempty"`
	Notes  []string `json:"notes,omitempty" dynamodbav:"notes,omitempty"`

	// Audit timestamps (epoch seconds)
	CreatedAt int64  `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt int64  `json:"updatedAt" dynamodbav:"updatedAt"`
	DeletedAt *int64 `json:"deletedAt,omitempty" dynamodbav:"deletedAt,omitempty"`

	// DynamoDB keys
	PK string `json:"-" dynamodbav:"PK"` // accountId
	SK string `json:"-" dynamodbav:"SK"` // {taskId}#{laborLineId}
}

// CreateLaborLineInput represents the input for creating a new labor line.
type CreateLaborLineInput struct {
	ContactID string   `json:"contactId"`
	AccountID string   `json:"accountId"`
	TaskID    string   `json:"taskId"`
	PartID    []string `json:"partId,omitempty"`
	Notes     []string `json:"notes,omitempty"`
}

// UpdateLaborLineInput represents the input for updating an existing labor line.
type UpdateLaborLineInput struct {
	LaborLineID string   `json:"laborLineId"`
	ContactID   string   `json:"contactId"`
	AccountID   string   `json:"accountId"`
	TaskID      string   `json:"taskId"`
	PartID      []string `json:"partId,omitempty"`
	Notes       []string `json:"notes,omitempty"`
}

// GetLaborLineInput represents the input for retrieving a labor line.
type GetLaborLineInput struct {
	AccountID   string `json:"accountId"`
	TaskID      string `json:"taskId"`
	LaborLineID string `json:"laborLineId"`
}

// ListLaborLinesInput represents the input for listing labor lines.
type ListLaborLinesInput struct {
	AccountID string `json:"accountId"`
	TaskID    string `json:"taskId,omitempty"` // Optional filter by task
}

// DeleteLaborLineInput represents the input for deleting a labor line.
type DeleteLaborLineInput struct {
	AccountID   string `json:"accountId"`
	TaskID      string `json:"taskId"`
	LaborLineID string `json:"laborLineId"`
}

// NewLaborLine creates a new LaborLine from CreateLaborLineInput.
func NewLaborLine(input CreateLaborLineInput) *LaborLine {
	now := time.Now().Unix()
	laborLineID := uuid.New().String()

	return &LaborLine{
		LaborLineID: laborLineID,
		ContactID:   input.ContactID,
		AccountID:   input.AccountID,
		TaskID:      input.TaskID,
		PartID:      input.PartID,
		Notes:       input.Notes,
		CreatedAt:   now,
		UpdatedAt:   now,
		PK:          input.AccountID,
		SK:          input.TaskID + "#" + laborLineID,
	}
}

// ToLaborLine converts UpdateLaborLineInput to LaborLine for updates.
func (input UpdateLaborLineInput) ToLaborLine() *LaborLine {
	return &LaborLine{
		LaborLineID: input.LaborLineID,
		ContactID:   input.ContactID,
		AccountID:   input.AccountID,
		TaskID:      input.TaskID,
		PartID:      input.PartID,
		Notes:       input.Notes,
		UpdatedAt:   time.Now().Unix(),
		PK:          input.AccountID,
		SK:          input.TaskID + "#" + input.LaborLineID,
	}
}

// IsDeleted returns true if the labor line has been soft deleted.
func (ll *LaborLine) IsDeleted() bool {
	return ll.DeletedAt != nil
}

// SoftDelete marks the labor line as deleted with the current timestamp.
func (ll *LaborLine) SoftDelete() {
	now := time.Now().Unix()
	ll.DeletedAt = &now
	ll.UpdatedAt = now
}
