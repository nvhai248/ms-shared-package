package grpcclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"time"

	"github.com/nvhai248/ms-shared-package/common/ctx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

type ClientConfig struct {
	Target             string
	UseTLS             bool
	InsecureSkipVerify bool
	DialTimeout        time.Duration
	KeepAliveTime      time.Duration
	KeepAliveTimeout   time.Duration
}

func DefaultClientConfig(target string) ClientConfig {
	return ClientConfig{
		Target:           target,
		UseTLS:           false,
		DialTimeout:      5 * time.Second,
		KeepAliveTime:    10 * time.Second,
		KeepAliveTimeout: 3 * time.Second,
	}
}

// NewClientConn creates an optimized HTTP/2 gRPC client connection with keepalive, pooling, and tracing
func NewClientConn(cfg ClientConfig, extraOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	kacp := keepalive.ClientParameters{
		Time:                cfg.KeepAliveTime,
		Timeout:             cfg.KeepAliveTimeout,
		PermitWithoutStream: true,
	}

	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(kacp),
		grpc.WithChainUnaryInterceptor(
			ContextPropagationInterceptor(),
			LoggingClientInterceptor(),
		),
	}

	if cfg.UseTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	opts = append(opts, extraOpts...)

	dialCtx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, cfg.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC target %s: %w", cfg.Target, err)
	}

	return conn, nil
}

// ContextPropagationInterceptor automatically injects context headers (X-Request-ID, User, etc.) into outgoing gRPC metadata
func ContextPropagationInterceptor() grpc.UnaryClientInterceptor {
	return func(
		parentCtx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		outCtx := ctx.ToOutgoingGRPCContext(parentCtx)
		return invoker(outCtx, method, req, reply, cc, opts...)
	}
}

// LoggingClientInterceptor logs every outgoing gRPC request with duration and structured fields
func LoggingClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		parentCtx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()
		reqID := ctx.RequestID(parentCtx)

		err := invoker(parentCtx, method, req, reply, cc, opts...)
		duration := time.Since(start)

		if err != nil {
			slog.Error("gRPC client call failed",
				"method", method,
				"request_id", reqID,
				"duration_ms", duration.Milliseconds(),
				"error", err,
			)
		} else {
			slog.Info("gRPC client call completed",
				"method", method,
				"request_id", reqID,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return err
	}
}

// ServerLoggingInterceptor provides structured logging for gRPC server implementations
func ServerLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		serverCtx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		wrappedCtx := ctx.FromIncomingGRPCContext(serverCtx)
		reqID := ctx.RequestID(wrappedCtx)

		resp, err := handler(wrappedCtx, req)
		duration := time.Since(start)

		if err != nil {
			slog.Error("gRPC server request error",
				"method", info.FullMethod,
				"request_id", reqID,
				"duration_ms", duration.Milliseconds(),
				"error", err,
			)
		} else {
			slog.Info("gRPC server request completed",
				"method", info.FullMethod,
				"request_id", reqID,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return resp, err
	}
}

// TraceIDServerInterceptor extracts request-id from gRPC metadata into Context
func TraceIDServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		serverCtx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		var reqID string
		if md, ok := metadata.FromIncomingContext(serverCtx); ok {
			if vals := md.Get(ctx.RequestIDKey.String()); len(vals) > 0 {
				reqID = vals[0]
			}
		}
		wrappedCtx := ctx.ContextWithValues(serverCtx, reqID, "", "", "", "")
		return handler(wrappedCtx, req)
	}
}
