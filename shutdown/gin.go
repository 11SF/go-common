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
	"github.com/gin-gonic/gin"
)

const ginTag = "gin shutdown"

// Gin starts the Gin server and blocks until SIGINT/SIGTERM is received,
// then gracefully shuts down within the given timeout.
func Gin(ctx context.Context, router *gin.Engine, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return err
	}
	cfg.withDefaults()

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	go func() {
		logger.Info(ctx, "gin server starting", slog.String("addr", cfg.Addr), logger.LogAttrTag(ginTag))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(ctx, "gin server error", logger.LogAttrError(err), logger.LogAttrTag(ginTag))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(ctx, "gin server shutting down", logger.LogAttrTag(ginTag))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(ctx, "gin server forced to shutdown", logger.LogAttrError(err), logger.LogAttrTag(ginTag))
		return err
	}

	logger.Info(ctx, "gin server shutdown complete", logger.LogAttrTag(ginTag))
	return nil
}
