package cgroup

import "github.com/google/uuid"

type Message struct {
	Sequence int    `json:"sequence"`
	Payload  string `json:"payload"`
}

func NewMessage(seq int) Message {
	m, _ := uuid.NewRandom()
	return Message{
		Sequence: seq,
		Payload:  m.String(),
	}
}
