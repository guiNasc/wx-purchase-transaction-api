package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	server, cleanup, err := InitializeServer()
	if err != nil {
		os.Stderr.WriteString("failed to initialize application: " + err.Error() + "\n")
		os.Exit(1)
	}
	defer cleanup()

	slog.Default().Info("server starting", "addr", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Default().Error("server failed", "error", err)
		os.Exit(1)
	}

}
