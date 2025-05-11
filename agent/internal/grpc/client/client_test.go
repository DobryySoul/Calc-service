package client_test

import (
	"agent/internal/models/req"
	"context"
	"net"
	"testing"

	"agent/internal/grpc/client"
	pb "agent/pkg/api/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

type mockOrchestratorServer struct {
	pb.UnimplementedOrchestratorServiceServer
	getTaskHandler    func(context.Context, *emptypb.Empty) (*pb.Task, error)
	sendResultHandler func(context.Context, *pb.Result) (*emptypb.Empty, error)
}

func (m *mockOrchestratorServer) GetTask(ctx context.Context, req *emptypb.Empty) (*pb.Task, error) {
	return m.getTaskHandler(ctx, req)
}

func (m *mockOrchestratorServer) SendResult(ctx context.Context, res *pb.Result) (*emptypb.Empty, error) {
	return m.sendResultHandler(ctx, res)
}

func startMockServer(t *testing.T) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	return s, lis
}

func TestNewGRPCClient(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		s, lis := startMockServer(t)
		pb.RegisterOrchestratorServiceServer(s, &mockOrchestratorServer{})

		go func() {
			if err := s.Serve(lis); err != nil {
				t.Errorf("Server exited with error: %v", err)
			}
		}()
		defer s.Stop()

		conn, err := grpc.NewClient("bufnet",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return lis.Dial()
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err)
		defer conn.Close()

		logger := zap.NewNop()
		grpcClient, err := client.NewGRPCClient("bufnet", "", logger)
		require.NoError(t, err)
		assert.NotNil(t, grpcClient)
		assert.NoError(t, grpcClient.Close())
	})

	t.Run("Nil check", func(t *testing.T) {
		logger := zap.NewNop()
		_, err := client.NewGRPCClient("invalid", "99999", logger)
		assert.Nil(t, err)
	})
}

func TestGRPCClient_SendResult(t *testing.T) {
	tests := []struct {
		name        string
		result      req.Result
		userID      uint64
		mockHandler func(context.Context, *pb.Result) (*emptypb.Empty, error)
		wantErr     bool
	}{
		{
			name: "successful int result",
			result: req.Result{
				ID:    1,
				Value: 42,
			},
			userID: 123,
			mockHandler: func(ctx context.Context, res *pb.Result) (*emptypb.Empty, error) {
				assert.Equal(t, int32(1), res.Id)
				assert.Equal(t, uint64(123), res.UserId)
				assert.IsType(t, &pb.Result_IntResult{}, res.Value)
				return &emptypb.Empty{}, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, lis := startMockServer(t)
			pb.RegisterOrchestratorServiceServer(s, &mockOrchestratorServer{
				sendResultHandler: tt.mockHandler,
			})

			go func() {
				if err := s.Serve(lis); err != nil {
					t.Errorf("Server exited with error: %v", err)
				}
			}()
			defer s.Stop()

			conn, err := grpc.NewClient("bufnet",
				grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
					return lis.Dial()
				}),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			require.NoError(t, err)
			defer conn.Close()

			logger := zap.NewNop()
			grpcClient, err := client.NewGRPCClient("bufnet", "", logger)
			require.NoError(t, err)
			defer grpcClient.Close()

			grpcClient.SendResult(tt.result, tt.userID)
		})
	}
}

func TestGRPCClient_Close(t *testing.T) {
	s, lis := startMockServer(t)
	pb.RegisterOrchestratorServiceServer(s, &mockOrchestratorServer{})

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()
	defer s.Stop()

	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	logger := zap.NewNop()
	grpcClient, err := client.NewGRPCClient("bufnet", "", logger)
	require.NoError(t, err)

	assert.NoError(t, grpcClient.Close())
	assert.Error(t, grpcClient.Close(), "second close should return error")
}
