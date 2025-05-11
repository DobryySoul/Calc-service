package client

import (
	"agent/internal/models/req"
	"agent/internal/models/resp"
	"context"
	"fmt"
	"time"

	pb "agent/pkg/api/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client pb.OrchestratorServiceClient
	logger *zap.Logger
}

func NewGRPCClient(host string, port string, logger *zap.Logger) (*GRPCClient, error) {
	conn, err := grpc.NewClient(
		host+":"+port,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: pb.NewOrchestratorServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func (c *GRPCClient) GetTask() *resp.Task {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := c.client.GetTask(ctx, &emptypb.Empty{})
	if err != nil {
		c.logger.Error("error while getting task", zap.Error(err))
		return nil
	}

	var opTime time.Duration
	if response.OperationTime != nil {
		opTime = response.OperationTime.AsDuration()
	}

	return &resp.Task{
		ID:            int(response.Id),
		Arg1:          response.Arg1,
		Arg2:          response.Arg2,
		Operation:     response.Operation,
		OperationTime: opTime,
		UserID:        response.UserId,
	}
}

func (c *GRPCClient) SendResult(result req.Result, userID uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcResult := &pb.Result{
		Id:     int32(result.ID),
		UserId: userID,
	}

	switch v := result.Value.(type) {
	case int:
		grpcResult.Value = &pb.Result_IntResult{IntResult: int64(v)}
	case float64:
		grpcResult.Value = &pb.Result_FloatResult{FloatResult: v}
	case error:
		grpcResult.Value = &pb.Result_Error{Error: v.Error()}
	default:
		c.logger.Error("unsupported result type", zap.Any("type", v))
		return
	}

	_, err := c.client.SendResult(ctx, grpcResult)
	if err != nil {
		c.logger.Error("error while sending result", zap.Error(err))
	}
}
