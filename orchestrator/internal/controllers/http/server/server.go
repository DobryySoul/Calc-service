package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/DobryySoul/orchestrator/internal/config"
	server "github.com/DobryySoul/orchestrator/internal/controllers/grpc"
	"github.com/DobryySoul/orchestrator/internal/controllers/http/handler"
	"github.com/DobryySoul/orchestrator/internal/repository"
	"github.com/DobryySoul/orchestrator/internal/service"
	pb "github.com/DobryySoul/orchestrator/pkg/api/v1"
	"github.com/DobryySoul/orchestrator/pkg/middleware"
	posetgres "github.com/DobryySoul/orchestrator/pkg/postgres"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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

	workDir, _ := os.Getwd()
	frontendDir := filepath.Join(workDir, "frontend")

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
	})

	fs := http.FileServer(http.Dir(frontendDir))
	r.Handle("/*", fs)

	r.Route("/api/v1", func(r chi.Router) {

		r.Get("/register", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(frontendDir, "register.html"))
		})
		r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(frontendDir, "login.html"))
		})

		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)

		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(cfg.JWTConfig.Secret, logger))
			r.Post("/calculate", calcHandler.Calculate)
			r.Get("/expressions", calcHandler.ListAll)
			r.Get("/expressions/{id}", calcHandler.ListByID)

			// метод для удобства отслеживания количества операций и подведения статистики для frontend
			r.Get("/statistics", calcHandler.GetStatistics)
		})
	})

	r.Get("/internal/task", calcHandler.SendTask)
	r.Post("/internal/task", calcHandler.ReceiveResult)

	httpServer := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: r,
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second,
			MaxConnectionAge:      30 * time.Second,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  10 * time.Second,
			Timeout:               3 * time.Second,
		}),
	)

	orchestrator := server.NewGRPCServer(calcService)

	pb.RegisterOrchestratorServiceServer(grpcServer, orchestrator)

	errChan := make(chan error, 2)

	go func() {
		logger.Info("Starting HTTP server",
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port))

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
			errChan <- err
		}
	}()

	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			logger.Error("Failed to listen gRPC port",
				zap.String("port", cfg.GRPCPort),
				zap.Error(err))
			errChan <- fmt.Errorf("gRPC listen error: %w", err)
			return
		}

		logger.Info("Starting gRPC server",
			zap.String("port", cfg.GRPCPort))

		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
			errChan <- fmt.Errorf("gRPC serve error: %w", err)
		}
	}()

	shutdownFunc := func(ctx context.Context) error {
		logger.Info("Shutting down servers...")

		var errs []error

		if err := httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("HTTP server shutdown error: %w", err))
		}

		done := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-ctx.Done():
			grpcServer.Stop()
			errs = append(errs, fmt.Errorf("gRPC server forced to shutdown: %w", ctx.Err()))
		case <-done:
		}

		if len(errs) > 0 {
			return fmt.Errorf("shutdown completed with errors: %v", errs)
		}

		logger.Info("Servers stopped gracefully")
		return nil
	}

	select {
	case err := <-errChan:
		return shutdownFunc, fmt.Errorf("server startup error: %w", err)
	default:
		return shutdownFunc, nil
	}
}
