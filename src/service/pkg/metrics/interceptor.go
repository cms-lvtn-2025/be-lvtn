package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a new unary server interceptor that records metrics
func UnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Increment active connections
		IncActiveConnections(serviceName)
		defer DecActiveConnections(serviceName)

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Determine status
		grpcStatus := "success"
		if err != nil {
			grpcStatus = status.Code(err).String()
		}

		// Record metrics
		RecordGrpcRequest(serviceName, info.FullMethod, grpcStatus, duration)

		return resp, err
	}
}

// StreamServerInterceptor returns a new stream server interceptor that records metrics
func StreamServerInterceptor(serviceName string) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Increment active connections
		IncActiveConnections(serviceName)
		defer DecActiveConnections(serviceName)

		// Call the handler
		err := handler(srv, stream)

		// Calculate duration
		duration := time.Since(start)

		// Determine status
		grpcStatus := "success"
		if err != nil {
			grpcStatus = status.Code(err).String()
		}

		// Record metrics
		RecordGrpcRequest(serviceName, info.FullMethod, grpcStatus, duration)

		return err
	}
}
