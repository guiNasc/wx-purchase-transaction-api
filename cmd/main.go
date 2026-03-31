package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

const shutdownTimeout = 10 * time.Second

func main() {
	os.Exit(run())
}

func run() int {
	_ = godotenv.Load()
	server, cleanup, err := InitializeServer()
	if err != nil {
		os.Stderr.WriteString("failed to initialize application: " + err.Error() + "\n")
		return 1
	}
	defer cleanup()

	serverErr := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}

		close(serverErr)
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Default().Info("server starting", "addr", server.Addr)

	select {
	case err := <-serverErr:
		if err != nil {
			slog.Default().Error("server failed", "error", err)
			return 1
		}

		return 0
	case <-signalCtx.Done():
		slog.Default().Info("shutdown signal received", "signal", signalCtx.Err())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Default().Error("graceful shutdown failed", "error", err)
		if closeErr := server.Close(); closeErr != nil {
			slog.Default().Error("server close failed", "error", closeErr)
		}
		return 1
	}

	if err := <-serverErr; err != nil {
		slog.Default().Error("server shutdown finished with error", "error", err)
		return 1
	}

	slog.Default().Info("server stopped")

	return 0
}
