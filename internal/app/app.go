package app

import (
	"app/internal/api/v1"
	"context"
	"net"
	"net/http"
)

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
	listener    net.Listener
}

func NewApp(ctx context.Context) (*App, error) {
	app := &App{}

	if err := app.initDeps(ctx); err != nil {
		return nil, err
	}
	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	return app.httpServer.Serve(app.listener)
}

func (app *App) initDeps(ctx context.Context) error {
	funcs := []func(ctx context.Context) error{
		app.initDi,
		app.initListener,
		app.initHTTPServer,
	}

	for _, f := range funcs {
		if err := f(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) initDi(ctx context.Context) error {
	app.diContainer = NewDIContainer()
	return nil
}

func (app *App) initListener(ctx context.Context) error {
	addr := app.diContainer.cfg.HTTP.Addr
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	app.listener = listener
	return nil
}

func (app *App) initHTTPServer(ctx context.Context) error {
	orderSvc := app.diContainer.OrderService(ctx)

	api := v1.NewAPI(orderSvc)

	app.httpServer = &http.Server{
		Addr:    app.listener.Addr().String(),
		Handler: api.Router,
	}

	return nil
}

// DIContainer возвращает DI контейнер приложения
func (app *App) DIContainer() *diContainer {
	return app.diContainer
}

// Shutdown выполняет graceful shutdown HTTP сервера
func (app *App) Shutdown(ctx context.Context) error {
	if app.httpServer != nil {
		return app.httpServer.Shutdown(ctx)
	}
	return nil
}
