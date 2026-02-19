package worker

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const apiKeyHeader = "x-api-key"

// UnaryAuthInterceptor validates the x-api-key metadata on unary RPCs.
func UnaryAuthInterceptor(apiKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := validateAPIKey(ctx, apiKey); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamAuthInterceptor validates the x-api-key metadata on streaming RPCs.
func StreamAuthInterceptor(apiKey string) grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := validateAPIKey(ss.Context(), apiKey); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

func validateAPIKey(ctx context.Context, expected string) error {
	if expected == "" {
		return nil // no auth configured
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(apiKeyHeader)
	if len(values) == 0 {
		return status.Error(codes.Unauthenticated, "missing api key")
	}

	if values[0] != expected {
		return status.Error(codes.Unauthenticated, "invalid api key")
	}

	return nil
}
