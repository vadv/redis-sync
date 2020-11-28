package redis_sync

import (
	"context"
)

type Message struct {
	Key   string
	Value string
	TTL   string // "0" means ttl is not set
}

type Redis interface {
	RedisScanner
	RedisWriter
}

type RedisScanner interface {
	Scan(ctx context.Context, out chan<- *Message) error
}

type RedisWriter interface {
	Write(ctx context.Context, in <-chan *Message) error
}
