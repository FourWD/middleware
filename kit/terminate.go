package kit

import (
	"log/slog"
	"os"
)

func Terminate(logger *slog.Logger) {
	logger.Error("TERMINATE", "message", "application terminated")
	os.Exit(1)
}
