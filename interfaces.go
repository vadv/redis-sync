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
	RunScan(ctx context.Context) error
	ScanMessages() <-chan *Message
}

type RedisWriter interface {
	Write(ctx context.Context, in <-chan *Message) error
}
