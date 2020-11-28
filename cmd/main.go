package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	schema "gitlab.diskarte.net/engineering/redis-sync"
	"gitlab.diskarte.net/engineering/redis-sync/internal/redis"
)

var (
	fSource = flag.String("source", "redis://localhost:6379/0", "Connection to source redis")
	fOut    = flag.String("out", "redis://localhost:6379/1", "Connection to out redis")
)

func main() {

	flag.Parse()
	s, err := redis.New(*fSource)
	if err != nil {
		panic(err)
	}
	o, err := redis.New(*fOut)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	group, ctx := errgroup.WithContext(ctx)

	// stop signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	out, in := make(chan *schema.Message), s.ScanMessages()
	group.Go(func() error {
		if err := s.RunScan(ctx); err != nil {
			return err
		}
		log.Printf("[INFO] scan done\n")
		close(out)
		return ctx.Err()
	})

	group.Go(func() error {
		return o.Write(ctx, out)
	})

	group.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case m := <-in:
				log.Printf("[DEBUG] process messages: %#v\n", m)
				out <- m
			}
		}
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

	if err := group.Wait(); !errors.Is(err, context.Canceled) {
		log.Printf("[ERROR] shutdown: %s\n", err.Error())
	}

	log.Printf("[INFO] exit now\n")
}
