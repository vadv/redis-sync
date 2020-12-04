package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cheggaaa/pb/v3"
	"golang.org/x/sync/errgroup"

	schema "gitlab.diskarte.net/engineering/redis-sync"
	"gitlab.diskarte.net/engineering/redis-sync/internal/db"
	"gitlab.diskarte.net/engineering/redis-sync/internal/null"
	"gitlab.diskarte.net/engineering/redis-sync/internal/redis"
)

var (
	fSource           = flag.String("source", "redis://localhost:6379/0", "Connection to source redis")
	fOut              = flag.String("out", "redis://localhost:6379/1", "Connection to out redis")
	fLocalDB          = flag.String("db", "cproto://127.0.0.1:6534/redis-cache", "Reindexer dsn")
	fLocalDBNameSpace = flag.String("db-namespace", "default", "Namespace")
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
	case *fOut == `/dev/null`:
		o = null.Writer()
	default:
		o, err = redis.New(*fOut)
		if err != nil {
			panic(err)
		}
	}

	r, err := db.Open(*fLocalDB, *fLocalDBNameSpace)
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

	in, out := make(chan string, 64*1024), make(chan *schema.Message, 64*1024)
	group.Go(func() error {
		if err := s.ScanKeys(ctx, in); err != nil {
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
	bar := pb.Full.Start(count)

	group.Go(func() error {
		defer bar.Finish()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case key, ok := <-in:
				if !ok {
					close(out)
					return nil
				}
				m, ok := r.Get(key)
				if !ok {
					sourceM, err := s.Get(key)
					if err != nil {
						log.Printf("[ERROR] get key %#v: %s\n", key, err)
						break
					}
					if sourceM.TTL != "0" {
						if err := r.Set(sourceM); err != nil {
							panic(err)
						}
						bar.Increment()
						break
					}
					if err := r.Set(sourceM); err != nil {
						panic(err)
					}
					m = sourceM
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
