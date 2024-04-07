package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/rovilay/auth-service/app"
	"github.com/rovilay/auth-service/config"
	"github.com/rovilay/auth-service/repository"
	"github.com/rs/zerolog"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Str("application", "auth-service:main").Timestamp().Logger()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// add logger to context
	ctx = logger.WithContext(ctx)

	// load env
	envPath, err := filepath.Abs("./.env")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error resolving .env path")
	}

	err = godotenv.Load(envPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error loading .env file")
	}

	// load config
	c := config.LoadConfig(&logger)

	db, err := sqlx.ConnectContext(ctx, "pgx", c.DATABASE_URL)
	if err != nil {
		logger.Fatal().Err(err).Msg(fmt.Sprintf("failed to connect to DB %s", c.DATABASE_URL))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Err(err).Msg("failed to close postgres")
		}
	}()

	repo := repository.NewPostgresRepository(ctx, db, &logger)
	app := app.NewApp(repo, &c, &logger)

	if err = app.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to start app")
	}
}
