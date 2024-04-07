package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rovilay/auth-service/config"
	"github.com/rovilay/auth-service/repository"
	"github.com/rs/zerolog"
)

type App struct {
	router http.Handler
	config *config.AppConfig
	log    *zerolog.Logger
	repo   repository.UserRepository
}

func NewApp(repo repository.UserRepository, c *config.AppConfig, log *zerolog.Logger) *App {
	logger := log.With().Str("package:app", "App").Logger()

	app := &App{
		log:    &logger,
		config: c,
		repo:   repo,
	}

	app.loadRoutes()

	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.config.ServerPort),
		Handler: a.router,
	}

	a.log.Println("starting server on port: ", a.config.ServerPort)

	ch := make(chan error, 1)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		// wait 10 seconds before server shutdown
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return server.Shutdown(timeout)
	}
}
