package server

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/DobryySoul/orchestrator/internal/service"
	pb "github.com/DobryySoul/orchestrator/pkg/api/v1"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrchestratorServer struct {
	pb.UnimplementedOrchestratorServiceServer
	calcService *service.CalcService
	logger      *zap.Logger
}

type ServerOption func(*OrchestratorServer)

func NewGRPCServer(calcService *service.CalcService, opts ...ServerOption) *OrchestratorServer {
	server := &OrchestratorServer{
		calcService: calcService,
		logger:      zap.NewExample(),
	}

	return server
}

func RunServer(grpcServer *grpc.Server, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	log.Printf("gRPC server listening on :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *OrchestratorServer) GetTask(ctx context.Context, _ *emptypb.Empty) (*pb.Task, error) {
	task, err := s.calcService.GetTask(ctx, &emptypb.Empty{})
	if err != nil {
		s.logger.Info("GetTask failed:", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get task: %v", err)
	}

	if task == nil {
		return nil, status.Error(codes.NotFound, "no tasks available")
	}

	return task, nil
}

func (s *OrchestratorServer) SendResult(ctx context.Context, res *pb.Result) (*emptypb.Empty, error) {
	if res == nil {
		return nil, status.Error(codes.InvalidArgument, "result cannot be nil")
	}

	_, err := s.calcService.SendResult(ctx, res)
	if err != nil {
		s.logger.Info("SendResult failed for task %d: %v", zap.Int32("task_id", res.Id), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to store result: %v", err)
	}

	return &emptypb.Empty{}, nil
}
