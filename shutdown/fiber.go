package shutdown

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/11SF/go-common/logger"
	"github.com/gofiber/fiber/v2"
)

const fiberTag = "fiber shutdown"

// Fiber starts the Fiber server and blocks until SIGINT/SIGTERM is received,
// then gracefully shuts down within the given timeout.
func Fiber(ctx context.Context, app *fiber.App, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return err
	}
	cfg.withDefaults()

	go func() {
		logger.Info(ctx, "fiber server starting", slog.String("addr", cfg.Addr), logger.LogAttrTag(fiberTag))
		if err := app.Listen(cfg.Addr); err != nil {
			logger.Error(ctx, "fiber server error", logger.LogAttrError(err), logger.LogAttrTag(fiberTag))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "fiber server shutting down", logger.LogAttrTag(fiberTag))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error(ctx, "fiber server forced to shutdown", logger.LogAttrError(err), logger.LogAttrTag(fiberTag))
		return err
	}

	logger.Info(ctx, "fiber server shutdown complete", logger.LogAttrTag(fiberTag))
	return nil
}
