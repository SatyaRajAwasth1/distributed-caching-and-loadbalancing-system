package caching

import "time"

type Command string

const (
	CMDSet Command = "SET"
	CMDGet Command = "GET"
)

type MessageSet struct {
	Key   []byte
	Value []byte
	TTL   time.Duration
}

type MessageGet struct {
	Key []byte
}

type Message struct {
	Cmd   Command
	Key   []byte
	Value []byte
	TTl   time.Duration
}
