package redis

import (
	"context"

	"github.com/mediocregopher/radix/v3"

	schema "gitlab.diskarte.net/engineering/redis-sync"
)

type redis struct {
	pool *radix.Pool
}

func New(connection string) (schema.Redis, error) {
	db, err := radix.NewPool("tcp", connection, 0)
	return &redis{pool: db}, err
}

func (r *redis) getTTL(key string) (string, error) {
	var ttl string

	// Try getting key TTL.
	err := r.pool.Do(radix.Cmd(&ttl, "PTTL", key))
	if err != nil {
		return ttl, err
	}

	// When key has no expire PTTL returns "-1".
	// We set it to 0, default for no expiration time.
	if ttl == "-1" {
		ttl = "0"
	}

	return ttl, nil
}

func (r *redis) Scan(ctx context.Context, out chan<- *schema.Message) error {
	var key, value string
	scanner := radix.NewScanner(r.pool, radix.ScanAllKeys)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if !scanner.Next(&key) {
				return scanner.Close()
			}
			if err := r.pool.Do(radix.Cmd(&value, "DUMP", key)); err != nil {
				return err
			}
			ttl, err := r.getTTL(key)
			if err != nil {
				return err
			}
			out <- &schema.Message{Key: key, Value: value, TTL: ttl}
		}
	}
}

func (r *redis) Write(ctx context.Context, in <-chan *schema.Message) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-in:
			if !ok {
				return nil
			}
			if err := r.pool.Do(radix.Cmd(nil, "RESTORE", m.Key, m.TTL, m.Value, "REPLACE")); err != nil {
				return err
			}
		}
	}
}
