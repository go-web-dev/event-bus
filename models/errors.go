package models

type OperationRequestError struct {
	Body map[string]string `json:"body"`
}

func (e OperationRequestError) Error() string {
	return "missing required fields"
}

type InvalidJSONError struct {
}

func (e InvalidJSONError) Error() string {
	return "invalid json provided"
}
