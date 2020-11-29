package redis_sync

import (
	"context"
)

type Message struct {
	Key   string
	Value string
	TTL   string // "0" means ttl is not set
}

type Scanner interface {
	Scan(ctx context.Context, out chan<- *Message) error
}

type Writer interface {
	Write(ctx context.Context, in <-chan *Message) error
}

type Info interface {
	Size() (int, error)
}

type Redis interface {
	Scanner
	Writer
	Info
}
