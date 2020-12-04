package redis_sync

import (
	"context"
)

type Message struct {
	Key   string `reindex:"id,,pk"`
	Value string `reindex:"value"`
	TTL   string `reindex:"ttl"` // "0" means ttl is not set
}

type Scanner interface {
	ScanKeys(ctx context.Context, out chan<- string) error
	Get(key string) (*Message, error)
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

type DB interface {
	Get(key string) (*Message, bool)
	Set(value *Message) error
}
