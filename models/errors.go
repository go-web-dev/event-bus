package models

type RequiredField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type OperationRequestError struct {
	Body []RequiredField `json:"body"`
}

func (e OperationRequestError) Error() string {
	return "missing required fields"
}

type InvalidJSONError struct {
}

func (e InvalidJSONError) Error() string {
	return "invalid json provided"
}

type AuthError struct {
}

func (e AuthError) Error() string {
	return "unauthorized to make request"
}
