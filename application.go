package bootstrap

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/smartwalle/routine"
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

func WithBeforeStart(handler func(ctx context.Context)) Option {
	return func(app *Application) {
		app.beforeStart = handler
	}
}

func WithAfterStop(handler func(ctx context.Context)) Option {
	return func(app *Application) {
		app.afterStop = handler
	}
}

type Application struct {
	ctx    context.Context
	cancel func()

	stopTimeout time.Duration
	servers     []Server

	beforeStart func(context.Context)
	afterStop   func(context.Context)
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
	var rPool = routine.New(
		routine.WithCapacity(int32(len(app.servers)*2)+1),
		routine.WithMaxWorkers(int32(len(app.servers)*2)+1),
		routine.WithIdleTimeout(-1),
	)
	defer rPool.Close()
	rGroup, _ := rPool.Group(app.ctx)

	defer func() {
		if app.afterStop != nil {
			app.afterStop(context.WithoutCancel(rGroup.Context()))
		}
	}()

	if app.beforeStart != nil {
		app.beforeStart(context.WithoutCancel(rGroup.Context()))
	}

	for _, server := range app.servers {
		var nServer = server
		rGroup.Go(func(ctx context.Context) error {
			<-ctx.Done()
			stopCtx, stopCancel := context.WithTimeout(context.WithoutCancel(ctx), app.stopTimeout)
			defer stopCancel()
			return nServer.Stop(stopCtx)
		})

		rGroup.Go(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return nil
			default:
				return nServer.Start(ctx)
			}
		})
	}

	var sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	rGroup.Go(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		case <-sigs:
			return app.Shutdown()
		}
	})
	if err = rGroup.Wait(); err != nil && !errors.Is(err, context.Canceled) {
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
