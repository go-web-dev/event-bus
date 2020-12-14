package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/transport"
)

type streamSnapshotter interface {
	SnapshotStream(streamName string, w io.Writer) error
}

type snapshotStreamRequest struct {
	StreamName string `json:"stream_name" type:"string"`
}

type snapshotStreamResponse struct {
	Out []byte `json:"out" type:"[]byte]"`
}

func (r snapshotStreamResponse) Write(out []byte) (int, error) {
	// out is a proto.Message
	r.Out = out
	return 0, nil
}

func (router Router) snapshotStream(bus streamSnapshotter) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body snapshotStreamRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendError(w, snapshotStreamOperation, err)
			return err
		}

		var res snapshotStreamResponse
		err = bus.SnapshotStream(body.StreamName, res)
		if err != nil {
			transport.SendError(w, snapshotStreamOperation, err)
			return err
		}
		transport.SendJSON(w, snapshotStreamOperation, nil)
		return nil
	}
}
