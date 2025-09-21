package grpc

import (
	"context"
	"net"
	"strconv"

	"github.com/cmrd-a/shortener/internal/server/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor provides authentication for gRPC requests
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {

			ctx = context.WithValue(ctx, middleware.UserIDKey, middleware.GenerateUserID())
			return handler(ctx, req)
		}

		userIDHeader := md.Get("user_id")
		if len(userIDHeader) > 0 {
			if userID, err := strconv.ParseInt(userIDHeader[0], 10, 64); err == nil {
				ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
				return handler(ctx, req)
			}
		}

		tokens := md.Get("authorization")
		if len(tokens) > 0 {
			token := tokens[0]
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			if userID, err := middleware.ValidateJWT(token); err == nil {
				ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
				return handler(ctx, req)
			}
		}

		ctx = context.WithValue(ctx, middleware.UserIDKey, middleware.GenerateUserID())
		return handler(ctx, req)
	}
}

// TrustedSubnetInterceptor checks if the request comes from a trusted subnet
func TrustedSubnetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if info.FullMethod != "/shortener.ShortenerService/GetStats" {
			return handler(ctx, req)
		}

		if trustedSubnet == "" {
			return nil, status.Error(codes.PermissionDenied, "Access denied")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "No metadata provided")
		}

		realIPs := md.Get("x-real-ip")
		if len(realIPs) == 0 {
			return nil, status.Error(codes.PermissionDenied, "X-Real-IP header required")
		}

		clientIP := realIPs[0]

		_, trustedNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, "Invalid trusted subnet configuration")
		}

		ip := net.ParseIP(clientIP)
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "Invalid client IP")
		}

		if !trustedNet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
