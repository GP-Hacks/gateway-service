package grpc_clients

import (
	"fmt"

	proto "github.com/GP-Hacks/proto/pkg/api/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupChatClient(address string) (proto.ChatServiceClient, error) {
	// log.Debug("Attempting to create gRPC connection", slog.String("address", address))

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// log.Error("Failed to create gRPC connection", slog.String("address", address), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create gRPC connection with chat service: %w", err)
	}
	defer func() {
		if err != nil {
			_ = conn.Close()
			// log.Info("Closed gRPC connection due to error", slog.String("address", address))
		}
	}()

	chatClient := proto.NewChatServiceClient(conn)

	// log.Info("Successfully connected to chat service", slog.String("address", address))
	return chatClient, nil
}
