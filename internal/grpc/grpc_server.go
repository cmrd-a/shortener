package grpc

import (
	"net"

	pb "github.com/cmrd-a/shortener/api/proto/shortener"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// StartGRPCServer creates and starts a gRPC server
func StartGRPCServer(address string, service Servicer, trustedSubnet string, logger *zap.Logger) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			AuthInterceptor(),
			TrustedSubnetInterceptor(trustedSubnet),
		),
	)

	shortenerServer := NewServer(service)
	pb.RegisterShortenerServiceServer(grpcServer, shortenerServer)

	logger.Info("Starting gRPC server", zap.String("address", address))

	return grpcServer.Serve(listener)
}
