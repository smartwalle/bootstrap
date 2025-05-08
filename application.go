package bootstrap

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

type Option func(app *Application)

func WithContext(ctx context.Context) Option {
	return func(app *Application) {
		app.ctx = ctx
	}
}

func WithServers(servers ...Server) Option {
	return func(app *Application) {
		if len(servers) > 0 {
			app.servers = append(app.servers, servers...)
		}
	}
}

func WithStopTimeout(timeout time.Duration) Option {
	return func(app *Application) {
		app.stopTimeout = timeout
	}
}

type Application struct {
	ctx    context.Context
	cancel func()

	stopTimeout time.Duration
	servers     []Server
}

func New(opts ...Option) *Application {
	var app = &Application{}
	app.ctx = context.Background()
	app.stopTimeout = 10 * time.Second
	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}
	app.ctx, app.cancel = context.WithCancel(app.ctx)
	return app
}

func (app *Application) Run() (err error) {
	group, ctx := errgroup.WithContext(app.ctx)
	var wg = sync.WaitGroup{}

	for _, server := range app.servers {
		var nServer = server
		group.Go(func() error {
			<-ctx.Done()
			stopCtx, stopCancel := context.WithTimeout(context.WithoutCancel(ctx), app.stopTimeout)
			defer stopCancel()
			return nServer.Stop(stopCtx)
		})

		wg.Add(1)

		group.Go(func() error {
			wg.Done()
			select {
			case <-ctx.Done():
				return nil
			default:
				return nServer.Start(ctx)
			}
		})
	}

	wg.Wait()

	var sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	group.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-sigs:
			return app.Shutdown()
		}
	})
	if err = group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (app *Application) Shutdown() (err error) {
	if app.cancel != nil {
		app.cancel()
	}
	return nil
}
