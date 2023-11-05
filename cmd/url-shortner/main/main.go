package main

import (
	"log/slog"
	"os"
	"url-shortner/internal/config"
	"url-shortner/internal/lib/logger/sl"
	"url-shortner/internal/storage/client/postgres"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log = log.With(slog.String("env", cfg.Env))
	log.Info("starting url-shortner")
	log.Debug("debug enabled")

	/*
		for postgres db
	*/
	dbURL := "user=postgres dbname=url-shortner password=password sslmode=disable"
	storage, err := postgres.New(dbURL)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1) // or return
	}

	id, err := storage.SaveURL("https://www.youtube.com/", "youtube")
	if err != nil {
		log.Error("failed save url", sl.Err(err))
		os.Exit(1)
	}

	log.Info("saved url", slog.Int64("id", id))

	id, err = storage.SaveURL("https://www.youtube.com/", "youtube")
	if err != nil {
		log.Error("failed save url", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
