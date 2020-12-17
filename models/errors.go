package models

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
