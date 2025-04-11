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

type Application struct {
	ctx         context.Context
	cancel      func()
	stopTimeout time.Duration
	signals     []os.Signal

	services []Service
}

func New(servcies ...Service) *Application {
	ctx, cancel := context.WithCancel(context.Background())

	var app = &Application{}
	app.ctx = ctx
	app.cancel = cancel
	app.stopTimeout = 10 * time.Second
	app.signals = []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT}

	app.services = servcies

	return app
}

func (app *Application) Run() (err error) {
	var group, ctx = errgroup.WithContext(app.ctx)
	var wg = sync.WaitGroup{}

	for _, service := range app.services {
		var nService = service
		group.Go(func() error {
			<-ctx.Done()
			stopCtx, cancel := context.WithTimeout(app.ctx, app.stopTimeout)
			defer cancel()
			return nService.Stop(stopCtx)
		})
		wg.Add(1)
		group.Go(func() error {
			wg.Done()
			return nService.Start(app.ctx)
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
