package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"person-api/configs"
	"person-api/internal/handler"
	"person-api/internal/logger"
	"person-api/internal/services/enrichment"
	"person-api/internal/services/person"
	"person-api/internal/storage/postgres"

	_ "person-api/internal/handler/docs"
)

func main() {
	cfg, err := configs.LoadConfig()
	if err != nil {
		panic(err)
	}

	logg := logger.NewLogger(cfg.LogLevel)

	store, err := postgres.NewPostgresStorage(cfg.DBDSN)
	if err != nil {
		logg.Error("connect postgres", "err", err)
		os.Exit(1)
	}

	enrichSvc := enrichment.NewService()
	personSvc := person.NewPersonService(logg, enrichSvc, store)

	r := handler.NewRouter(personSvc)

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logg.Info("server started", "port", cfg.ServerPort)
		if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logg.Error("listen", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logg.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logg.Error("shutdown", "err", err)
	}
	logg.Info("stopped")
}
