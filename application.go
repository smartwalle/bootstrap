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

const (
	kStateIdle     = 0 // 未运行
	kStateRunning  = 1 // 运行中
	kStateFinished = 2 // 已结束
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
	if !app.state.CompareAndSwap(kStateIdle, kStateRunning) {
		switch app.state.Load() {
		case kStateRunning:
			return ErrApplicationRunning
		case kStateFinished:
			return ErrApplicationFinished
		default:
			return ErrBadApplication
		}
	}

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

func (app *Application) Stop() (err error) {
	//if !app.state.CompareAndSwap(kStateRunning, kStateFinished) {
	//	switch app.state.Load() {
	//	case kStateIdle:
	//		return ErrApplicationIdle
	//	case kStateFinished:
	//		return ErrApplicationFinished
	//	default:
	//  	var ErrApplicationIdle     = errors.New("application is idle")
	//		return ErrBadApplication
	//	}
	//}

	app.state.Store(kStateFinished)

	if app.cancel != nil {
		app.cancel()
	}
	return nil
}
