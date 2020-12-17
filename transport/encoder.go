package transport

import (
	"encoding/json"
	"io"

	"go.uber.org/zap"

	"github.com/go-web-dev/event-bus/logging"
	"github.com/go-web-dev/event-bus/models"
)

type response struct {
	Operation string      `json:"operation"`
	Status    bool        `json:"status"`
	Body      interface{} `json:"body,omitempty"`
	Context   interface{} `json:"context,omitempty"`
	Reason    string      `json:"reason,omitempty"`
}

// SendJSON is responsible for sending out JSON.
// To be used in successful cases only
func SendJSON(w io.Writer, op string, body interface{}) {
	logger := logging.Logger
	res := toResponse(body, op)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		logger.Error("could not encode json response", zap.Error(err))
	}
}

// SendError is responsible for sending out JSON.
// To be used in failure/negative cases only
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
