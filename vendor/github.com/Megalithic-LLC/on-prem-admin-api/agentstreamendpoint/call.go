package agentstreamendpoint

import (
	"github.com/Megalithic-LLC/on-prem-admin-api/agentstreamendpoint/emailproto"
)

type Call struct {
	Req   emailproto.ServerMessage
	Res   *emailproto.ClientMessage
	Done  chan bool
	Error error
}

func NewCall(req emailproto.ServerMessage) *Call {
	return &Call{
		Req:  req,
		Done: make(chan bool),
	}
}
