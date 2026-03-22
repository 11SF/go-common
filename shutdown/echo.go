package shutdown

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/11SF/go-common/logger"
	"github.com/labstack/echo/v4"
)

const echoTag = "echo shutdown"

// Echo starts the Echo server and blocks until SIGINT/SIGTERM is received,
// then gracefully shuts down within the given timeout.
func Echo(ctx context.Context, e *echo.Echo, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return err
	}
	cfg.withDefaults()

	go func() {
		logger.Info(ctx, "echo server starting", slog.String("addr", cfg.Addr), logger.LogAttrTag(echoTag))
		if err := e.Start(cfg.Addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "echo server error", logger.LogAttrError(err), logger.LogAttrTag(echoTag))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "echo server shutting down", logger.LogAttrTag(echoTag))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "echo server forced to shutdown", logger.LogAttrError(err), logger.LogAttrTag(echoTag))
		return err
	}

	logger.Info(ctx, "echo server shutdown complete", logger.LogAttrTag(echoTag))
	return nil
}
