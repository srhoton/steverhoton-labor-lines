package models

import "encoding/json"

// AppSyncEvent represents the structure of an AWS AppSync event.
type AppSyncEvent struct {
	TypeName   string                 `json:"typeName"`
	FieldName  string                 `json:"fieldName"`
	Arguments  map[string]interface{} `json:"arguments"`
	Identity   map[string]interface{} `json:"identity"`
	Source     map[string]interface{} `json:"source"`
	Request    AppSyncRequest         `json:"request"`
	Info       AppSyncInfo            `json:"info"`
	PrevResult map[string]interface{} `json:"prev"`
}

// AppSyncRequest contains request information.
type AppSyncRequest struct {
	Headers    map[string]string `json:"headers"`
	DomainName string            `json:"domainName"`
}

// AppSyncInfo contains information about the GraphQL operation.
type AppSyncInfo struct {
	FieldName        string                 `json:"fieldName"`
	ParentTypeName   string                 `json:"parentTypeName"`
	Variables        map[string]interface{} `json:"variables"`
	SelectionSetList []string               `json:"selectionSetList"`
}

// AppSyncResponse represents the response structure for AppSync.
type AppSyncResponse struct {
	Data  interface{}   `json:"data,omitempty"`
	Error *AppSyncError `json:"error,omitempty"`
}

// AppSyncError represents an error in AppSync format.
type AppSyncError struct {
	Message   string                 `json:"message"`
	Type      string                 `json:"type"`
	ErrorInfo map[string]interface{} `json:"errorInfo,omitempty"`
}

// GetArgumentAs extracts and unmarshals an argument from the AppSync event.
func (e *AppSyncEvent) GetArgumentAs(key string, target interface{}) error {
	if arg, exists := e.Arguments[key]; exists {
		// Convert to JSON and back to properly unmarshal into target type
		data, err := json.Marshal(arg)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, target)
	}
	return nil
}

// GetInputArgument extracts the 'input' argument and unmarshals it.
func (e *AppSyncEvent) GetInputArgument(target interface{}) error {
	return e.GetArgumentAs("input", target)
}
