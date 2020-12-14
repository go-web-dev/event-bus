package controllers

import (
	"io"

	"github.com/chill-and-code/event-bus/transport"
)

type dbSnapshotter interface {
	SnapshotDB(output string) error
}

type snapshotDBRequest struct {
	Output string `json:"output" type:"string"`
}

// snapshot individual stream
// don't write to server disk, let the client choose and write on its own
// just send down the protobuf stream

func (router Router) snapshotDB(bus dbSnapshotter) func(io.Writer, request) error {
	return func(w io.Writer, r request) error {
		var body snapshotDBRequest
		err := parseReq(r, &body)
		if err != nil {
			transport.SendJSON(w, snapshotDBOperation, err)
			return err
		}

		err = bus.SnapshotDB(body.Output)
		if err != nil {
			transport.SendJSON(w, snapshotDBOperation, err)
			return err
		}
		transport.SendJSON(w, snapshotDBOperation, nil)
		return nil
	}
}
