package bootstrap

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
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

type State int32

const (
	StateIdle     State = 0 // 未运行
	StateRunning  State = 1 // 运行中
	StateFinished State = 2 // 已结束
)

var (
	ErrApplicationRunning  = errors.New("applicatoin is running")
	ErrApplicationFinished = errors.New("application finished")
	ErrBadApplication      = errors.New("bad application")
)

type Application struct {
	ctx    context.Context
	cancel func()

	stopTimeout time.Duration
	servers     []Server
	state       atomic.Int32
}

func New(opts ...Option) *Application {
	var app = &Application{}
	app.ctx = context.Background()
	app.stopTimeout = 10 * time.Second
	app.state.Store(int32(StateIdle))
	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}
	app.ctx, app.cancel = context.WithCancel(app.ctx)
	return app
}

func (app *Application) Run() (err error) {
	if !app.state.CompareAndSwap(int32(StateIdle), int32(StateRunning)) {
		switch State(app.state.Load()) {
		case StateRunning:
			return ErrApplicationRunning
		case StateFinished:
			return ErrApplicationFinished
		default:
			return ErrBadApplication
		}
	}

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
			return app.Stop()
		}
	})
	if err = group.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

func (app *Application) State() State {
	return State(app.state.Load())
}

func (app *Application) Stop() (err error) {
	app.state.Store(int32(StateFinished))

	if app.cancel != nil {
		app.cancel()
	}
	return nil
}
