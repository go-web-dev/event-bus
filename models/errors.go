package models

import (
	"fmt"
	"strings"
)

// Error represents the Event Bus generic error
type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}

// RequiredField represents request required fields hint
type RequiredField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// OperationRequestError is returned in case the request has missing or mismatch fields
type OperationRequestError struct {
	Body []RequiredField `json:"body"`
}

func (e OperationRequestError) Error() string {
	return "missing required fields"
}

// OperationNotFoundError is returned in case the router switch has received an unknown operation
type OperationNotFoundError struct {
}

func (e OperationNotFoundError) Error() string {
	ops := []string{
		"health",
		"create_stream",
		"delete_stream",
		"get_stream_info",
		"get_stream_events",
		"write_event",
		"process_events",
		"retry_events",
		"retry_events",
		"mark_event",
		"exit",
	}
	return fmt.Sprintf(
		"operation must be one of: '%s'",
		strings.Join(ops, "', '"),
	)
}

// InvalidJSONError is returned in case there is any type of JSON errors
type InvalidJSONError struct {
}

func (e InvalidJSONError) Error() string {
	return "invalid json provided"
}

// AuthError is returned in case of any authentication/authorization errors
type AuthError struct {
}

func (e AuthError) Error() string {
	return "unauthorized to make request"
}
