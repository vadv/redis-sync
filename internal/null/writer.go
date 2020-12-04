package null

import (
	"context"

	schema "gitlab.diskarte.net/engineering/redis-sync"
)

func Writer() schema.Writer {
	return &w{}
}

type w struct{}

func (n *w) Write(ctx context.Context, in <-chan *schema.Message) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-in:
			if !ok {
				return nil
			}
		}
	}
}
