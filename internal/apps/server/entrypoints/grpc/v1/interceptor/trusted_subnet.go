package interceptor

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
)

func TrustedSubnet(logger *zap.Logger, subnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if subnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Debug("trusted subnet: missing metadata")
			return nil, status.Error(codes.PermissionDenied, "missing metadata")
		}
		values := md.Get("X-Real-IP")
		if len(values) == 0 {
			logger.Debug("trusted subnet: missing IP-address")
			return nil, status.Error(codes.PermissionDenied, "missing IP-address")
		}
		ipStr := values[0]
		ip := net.ParseIP(values[0])
		if ip == nil {
			logger.Debug("trusted subnet: invalid IP-address format", zap.String("ip", ipStr))
			return nil, status.Error(codes.PermissionDenied, "invalid IP-address format")
		}
		if !subnet.Contains(ip) {
			logger.Debug("trusted subnet: IP-address doesn't belong to the subnet", zap.String("ip", ipStr), zap.String("subnet", subnet.String()))
			return nil, status.Error(codes.PermissionDenied, "IP-address doesn't belong to the subnet")
		}

		return handler(ctx, req)
	}
}
