package grpc_clients

import (
	"fmt"
	"log/slog"

	proto "github.com/GP-Hacks/proto/pkg/api/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupUsersClient(address string, log *slog.Logger) (proto.UserServiceClient, error) {
	log.Debug("Attempting to create gRPC connection", slog.String("address", address))

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to create gRPC connection", slog.String("address", address), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create gRPC connection with charity service: %w", err)
	}
	defer func() {
		if err != nil {
			_ = conn.Close()
			log.Info("Closed gRPC connection due to error", slog.String("address", address))
		}
	}()

	charityClient := proto.NewUserServiceClient(conn)

	log.Info("Successfully connected to charity service", slog.String("address", address))
	return charityClient, nil
}
