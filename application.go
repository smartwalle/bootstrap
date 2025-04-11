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
		if ctx == nil {
			ctx = context.Background()
		}
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

	signals     []os.Signal
	stopTimeout time.Duration
	servers     []Server
}

func New(opts ...Option) *Application {
	var app = &Application{}
	app.ctx = context.Background()
	app.signals = []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}
	app.stopTimeout = 10 * time.Second

	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}

	ctx, cancel := context.WithCancel(app.ctx)
	app.ctx = ctx
	app.cancel = cancel
	return app
}

func (app *Application) Run() (err error) {
	var group, ctx = errgroup.WithContext(app.ctx)
	var wg = sync.WaitGroup{}

	for _, server := range app.servers {
		var nServer = server
		group.Go(func() error {
			<-ctx.Done()
			stopCtx, cancel := context.WithTimeout(context.WithoutCancel(app.ctx), app.stopTimeout)
			defer cancel()
			return nServer.Stop(stopCtx)
		})
		wg.Add(1)
		group.Go(func() error {
			wg.Done()
			return nServer.Start(app.ctx)
		})
	}

	wg.Wait()

	var c = make(chan os.Signal, 1)
	signal.Notify(c, app.signals...)

	group.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-c:
			return app.Stop()
		}
	})
	if err = group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (app *Application) Stop() (err error) {
	if app.cancel != nil {
		app.cancel()
	}
	return nil
}
