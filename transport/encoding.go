package transport

import (
	"encoding/json"
	"io"
	"log"

	"github.com/chill-and-code/event-bus/models"
)

type response struct {
	Operation string      `json:"operation"`
	Status    bool        `json:"status"`
	Body      interface{} `json:"body,omitempty"`
	Context   interface{} `json:"context,omitempty"`
	Reason    string      `json:"reason,omitempty"`
}

func SendJSON(w io.Writer, op string, body interface{}) {
	res := toResponse(body, op)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Fatal("could not send response:", err)
	}
}

func SendError(w io.Writer, op string, err error) {
	SendJSON(w, op, err)
}

func toResponse(any interface{}, op string) response {
	res := response{
		Operation: op,
	}
	switch value := any.(type) {
	case nil:
		res.Status = true
	case models.OperationRequestError:
		res.Status = false
		res.Reason = value.Error()
		res.Context = value
	case error:
		res.Status = false
		res.Reason = value.Error()
	default:
		res.Status = true
		res.Body = value
	}
	return res
}
