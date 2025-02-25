package server

import (
	"context"
	"fmt"
	"net/http"

	// "github.com/DobryySoul/Calc-service/internal/app/orchestrator/config"
	"github.com/DobryySoul/Calc-service/internal/configs"
	"github.com/DobryySoul/Calc-service/internal/http/handler"
	"github.com/DobryySoul/Calc-service/internal/service"
	"github.com/DobryySoul/Calc-service/pkg/middleware/logger"
	"go.uber.org/zap"
)

func Run(ctx context.Context, logger *zap.Logger, cfg configs.Config) (func(context.Context) error, error) {

	calcService := service.NewCalcService(cfg)

	muxHandler, err := newMuxHandler(ctx, logger, calcService)
	if err != nil {
		logger.Error("server initialization error", zap.Error(err))
		return nil, fmt.Errorf("server initialization error: %w", err)
	}

	// addr := os.Getenv("ADDR")
	// if addr == "" {
	// 	addr = "8080"
	// }

	// addr = ":" + addr
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

	muxHandler = handler.Middlewares(muxHandler, logger.LoggerMiddleware(log))

	return muxHandler, nil
}
