package null

import (
	"context"

	schema "gitlab.diskarte.net/engineering/redis-sync"
)

type Writer struct{}

func (n *Writer) Write(ctx context.Context, in <-chan *schema.Message) error {
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
