package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mucusscraper/Order-Flow/internal/logger"

	"github.com/mucusscraper/Order-Flow/internal/config"
)

func main() {
	cfg := config.Load()
	log := logger.New()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: mux,
	}
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		log.Info("starting server", "port", cfg.ServerPort, "env", cfg.AppEnv)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()
	sig := <-shutdownChan
	log.Info("shutdown signal received", "signal", sig.String())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Error("forced server shutdown", "error", err)
	}
	log.Info("server stopped gracefully")
}
