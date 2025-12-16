package app

import (
	"app/internal/closer"
	"app/internal/config"
	v1 "app/internal/http/v1"
	"app/internal/logger"
	"app/internal/otelx"
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"syscall"
)

const serviceName = "wb-orders"

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
	listener    net.Listener
	otel        otelx.InitResult
}

func NewApp(ctx context.Context) (*App, error) {
	app := &App{}
	if err := app.initDeps(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	if app.httpServer == nil || app.listener == nil {
		return errors.New("app not initialized")
	}
	err := app.httpServer.Serve(app.listener)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (app *App) initDeps(ctx context.Context) error {
	funcs := []func(context.Context) error{
		app.initConfig,
		app.initOTel,
		app.initLogger,
		app.initCloser,
		app.initDi,
		app.initInfra,
		app.initListener,
		app.initHTTPServer,
		app.registerClosers,
	}
	for _, f := range funcs {
		if err := f(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (app *App) initConfig(ctx context.Context) error {
	_ = ctx
	return config.Init()
}

func (app *App) initOTel(ctx context.Context) error {
	tel, err := otelx.Init(ctx, serviceName, config.AppConfig.Env)
	if err != nil {
		return err
	}
	app.otel = tel
	return nil
}

func (app *App) initLogger(ctx context.Context) error {
	_ = ctx
	return logger.Init(
		config.AppConfig.Logger.Level,
		config.AppConfig.Logger.AsJSON,
		app.otel.Providers.LoggerProvider,
	)
}

func (app *App) initCloser(ctx context.Context) error {
	_ = ctx
	closer.SetLogger(logger.Logger())
	closer.Configure(os.Interrupt, syscall.SIGTERM)
	return nil
}

func (app *App) initDi(ctx context.Context) error {
	_ = ctx
	app.diContainer = NewDIContainer()
	return nil
}

func (app *App) initInfra(ctx context.Context) error {
	return app.diContainer.Init(ctx)
}

func (app *App) initListener(ctx context.Context) error {
	_ = ctx
	addr := config.AppConfig.HTTP.Addr

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	app.listener = l
	return nil
}

func (app *App) initHTTPServer(ctx context.Context) error {
	svc, err := app.diContainer.OrderService(ctx)
	if err != nil {
		return err
	}

	api, err := v1.NewAPI(svc)
	if err != nil {
		return err
	}

	app.httpServer = &http.Server{
		Addr:    app.listener.Addr().String(),
		Handler: api,
	}
	return nil
}

func (app *App) registerClosers(ctx context.Context) error {
	_ = ctx

	closer.AddNamed("otel-shutdown", func(ctx context.Context) error {
		if app.otel.Shutdown == nil {
			return nil
		}
		return app.otel.Shutdown(ctx)
	})

	closer.AddNamed("http-server", func(ctx context.Context) error {
		if app.httpServer == nil {
			return nil
		}
		return app.httpServer.Shutdown(ctx)
	})

	closer.AddNamed("listener", func(ctx context.Context) error {
		if app.listener == nil {
			return nil
		}
		return app.listener.Close()
	})

	return nil
}

func (app *App) DIContainer() *diContainer {
	return app.diContainer
}
