package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/DobryySoul/orchestrator/internal/config"
	"github.com/DobryySoul/orchestrator/internal/http/handler"
	"github.com/DobryySoul/orchestrator/internal/repository"
	"github.com/DobryySoul/orchestrator/internal/service"
	"github.com/DobryySoul/orchestrator/pkg/middleware"
	posetgres "github.com/DobryySoul/orchestrator/pkg/postgres"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func Run(ctx context.Context, logger *zap.Logger, cfg *config.Config) (func(context.Context) error, error) {
	pg, err := posetgres.NewConn(ctx, &cfg.PostgresConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	authRepo := repository.NewAuthRepo(pg)

	calcService := service.NewCalcService(cfg, logger)
	authService := service.NewAuthService(authRepo, cfg.JWTConfig.Secret, cfg.JWTConfig.TTL, logger)

	r := chi.NewRouter()

	r.Use(
		middleware.LoggerMiddleware(logger),
		middleware.RecoveryMiddleware(logger),
		middleware.AllowCORS,
	)

	calcHandler := handler.NewCalcHandler(logger, calcService)
	authHandler := handler.NewAuthHandler(logger, authService)

	r.Route("/api/v1", func(r chi.Router) {

		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTConfig.Secret, logger))
			r.Post("/calculate", calcHandler.Calculate)
			r.Get("/expressions", calcHandler.ListAll)
			r.Get("/expressions/{id}", calcHandler.ListByID)

			// метод для удобства отслеживания количества операций и подведения статистики для frontend
			r.Get("/statistics", calcHandler.GetStatistics)

			r.Handle("/", http.FileServer(http.Dir(filepath.Join("frontend"))))

		})
	})
	r.Get("/internal/task", calcHandler.SendTask)
	r.Post("/internal/task", calcHandler.ReceiveResult)

	srv := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("Starting server",
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", zap.Error(err))
		}
	}()

	return srv.Shutdown, nil
}
