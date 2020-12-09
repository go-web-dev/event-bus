package models

type OperationRequestError struct {
	Fields map[string]string `json:"fields"`
}

func (e OperationRequestError) Error() string {
	return "missing required fields"
}
