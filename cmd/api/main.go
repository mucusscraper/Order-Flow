package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mucusscraper/Order-Flow/internal/config"
	"github.com/mucusscraper/Order-Flow/internal/database"
	"github.com/mucusscraper/Order-Flow/internal/handler"
	"github.com/mucusscraper/Order-Flow/internal/logger"
	"github.com/mucusscraper/Order-Flow/internal/middleware"
	"github.com/mucusscraper/Order-Flow/internal/repository"
	"github.com/mucusscraper/Order-Flow/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.New()
	ctx := context.Background()

	sqlDB, err := database.InitDB(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to initialize database and run migrations", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	dbPool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to create pgx pool", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	rdb, err := database.InitRedis(ctx, cfg.RedisURL)
	if err != nil {
		log.Error("failed to initialize redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()

	orderRepo := repository.NewPostgresOrderRepository(dbPool)
	orderService := service.NewOrderService(orderRepo, rdb)
	orderHandler := handler.NewOrderHandler(orderService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle("POST /orders", middleware.RateLimitMiddleware(rdb, 5, 1*time.Minute)(
		middleware.AuthMiddleware(cfg.JWTSecret)(
			middleware.RBACMiddleware("CUSTOMER", "ADMIN")(http.HandlerFunc(orderHandler.CreateOrder)),
		),
	))

	mux.Handle("GET /orders/{id}", middleware.RateLimitMiddleware(rdb, 5, 1*time.Minute)(
		middleware.AuthMiddleware(cfg.JWTSecret)(
			middleware.RBACMiddleware("CUSTOMER", "OPERATOR", "ADMIN")(http.HandlerFunc(orderHandler.GetOrder)),
		),
	))

	mux.Handle("POST /orders/{id}/cancel", middleware.RateLimitMiddleware(rdb, 5, 1*time.Minute)(
		middleware.AuthMiddleware(cfg.JWTSecret)(
			middleware.RBACMiddleware("CUSTOMER", "ADMIN")(http.HandlerFunc(orderHandler.CancelOrder)),
		),
	))

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		log.Error("forced server shutdown", "error", err)
	}
	log.Info("server stopped gracefully")
}
