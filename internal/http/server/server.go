package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DobryySoul/Calc-service/internal/config"
	"github.com/DobryySoul/Calc-service/internal/http/handler"
	"github.com/DobryySoul/Calc-service/internal/service"
	"github.com/DobryySoul/Calc-service/pkg/middleware"
	"go.uber.org/zap"
)

func Run(ctx context.Context, logger *zap.Logger, cfg config.Config) (func(context.Context) error, error) {

	calcService := service.NewCalcService(cfg, logger)

	muxHandler, err := newMuxHandler(ctx, logger, calcService)
	if err != nil {
		logger.Error("server initialization error", zap.Error(err))
		return nil, fmt.Errorf("server initialization error: %w", err)
	}

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: muxHandler}

	go func() {
		logger.Info("START SERVER", zap.String("port", cfg.Port))

		err := srv.ListenAndServe()
		if err != nil {
			logger.Error("server error", zap.Error(err))
		}
	}()

	return srv.Shutdown, nil
}

func newMuxHandler(ctx context.Context, log *zap.Logger, calcService *service.CalcService) (http.Handler, error) {
	muxHandler, err := handler.NewHandler(ctx, log, calcService)
	if err != nil {
		log.Error("handler initialization error", zap.Error(err))
		return nil, fmt.Errorf("handler initialization error: %w", err)
	}

	muxHandler = handler.Middlewares(muxHandler, middleware.RecoveryMiddleware(log))
	muxHandler = handler.Middlewares(muxHandler, middleware.LoggerMiddleware(log))
	muxHandler = middleware.AllowCORS(muxHandler)
	
	return muxHandler, nil
}
