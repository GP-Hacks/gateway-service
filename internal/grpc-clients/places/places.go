package grpc_clients

import (
	"context"
	"fmt"
	"time"

	"github.com/GP-Hacks/kdt2024-commons/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupPlacesClient(address string) (proto.PlacesServiceClient, error) {
	// log.Debug("Attempting to create gRPC connection", slog.String("address", address))

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// log.Error("Failed to create gRPC connection", slog.String("address", address), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create gRPC connection with places service: %w", err)
	}
	defer func() {
		if err != nil {
			_ = conn.Close()
			// log.Info("Closed gRPC connection due to error", slog.String("address", address))
		}
	}()

	placesClient := proto.NewPlacesServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// log.Debug("Performing health check on places service", slog.String("address", address))
	healthResponse, err := placesClient.HealthCheck(ctx, &proto.HealthCheckRequest{})
	if err != nil {
		// log.Error("Health check failed", slog.String("address", address), slog.String("error", err.Error()))
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	if !healthResponse.IsHealthy {
		err = fmt.Errorf("places service is not healthy")
		// log.Warn("Places service reported as unhealthy", slog.String("address", address))
		return nil, err
	}

	// log.Info("Successfully connected to places service", slog.String("address", address))
	return placesClient, nil
}
