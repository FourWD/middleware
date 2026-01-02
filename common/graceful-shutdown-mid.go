package common

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func GracefulShutdown(engineCancel context.CancelFunc) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case <-quit:
		Log("Received shutdown signal", nil, "")
	case err := <-serverErrChan:
		Log("HTTP server error", map[string]interface{}{"error": err.Error()}, "")
	}

	Log("Shutting down server...", nil, "")

	// Cancel engine context to stop all engines
	if engineCancel != nil {
		engineCancel()
		Log("Engine shutdown signal sent", nil, "")
	}

	// Give engines time to shutdown gracefully
	time.Sleep(2 * time.Second)

	// Shutdown HTTP server with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := fiberApp.ShutdownWithContext(ctx); err != nil {
		Log("Error shutting down HTTP server", map[string]interface{}{"error": err.Error()}, "")
	}

	// Cleanup database connection
	if Database != nil {
		if sqlDB, err := Database.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				Log("Error closing database connection", map[string]interface{}{"error": err.Error()}, "")
			} else {
				Log("Database connection closed", nil, "")
			}
		}
	}

	// Close Firebase client
	if FirebaseClient != nil {
		if err := FirebaseClient.Close(); err != nil {
			Log("Error closing Firebase client", map[string]interface{}{"error": err.Error()}, "")
		} else {
			Log("Firebase client closed", nil, "")
		}
	}

	Log("Server shutdown complete", nil, "")
}
