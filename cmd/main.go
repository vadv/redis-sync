package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/cheggaaa/pb/v3"
	"golang.org/x/sync/errgroup"

	schema "gitlab.diskarte.net/engineering/redis-sync"
	"gitlab.diskarte.net/engineering/redis-sync/internal/null"
	"gitlab.diskarte.net/engineering/redis-sync/internal/redis"
)

var (
	fSource = flag.String("source", "redis://localhost:6379/0", "Connection to source redis")
	fOut    = flag.String("out", "redis://localhost:6379/1", "Connection to out redis")
)

func main() {

	flag.Parse()

	var err error
	var s schema.Redis
	var o schema.Writer

	s, err = redis.New(*fSource)
	if err != nil {
		panic(err)
	}

	switch {
	case strings.HasPrefix(*fOut, `redis://`):
		o, err = redis.New(*fOut)
		if err != nil {
			panic(err)
		}
	default:
		o = &null.Writer{}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)

	// stop signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	in, out := make(chan *schema.Message, 1024), make(chan *schema.Message, 1024)
	group.Go(func() error {
		if err := s.Scan(ctx, in); err != nil {
			return err
		}
		log.Printf("[INFO] scan done\n")
		cancel()
		return ctx.Err()
	})

	count, err := s.Size()
	if err != nil {
		panic(err)
	}
	bar := pb.StartNew(count)

	group.Go(func() error {
		defer bar.Finish()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m, ok := <-in:
				if !ok {
					close(out)
					return nil
				}
				out <- m
				bar.Increment()
			}
		}
	})

	group.Go(func() error {
		return o.Write(ctx, out)
	})

	// handle termination
	select {
	case <-sig:
		log.Printf("[INFO] shutdown signal received\n")
		cancel()
		break
	case <-ctx.Done():
		break
	}

	if err := group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("[ERROR] shutdown: %s\n", err.Error())
	}

	log.Printf("[INFO] exit now\n")
}
