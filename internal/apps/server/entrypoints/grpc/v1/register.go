package v1

import (
	"context"
	"fmt"
	"github.com/dlomanov/mon/internal/apps/server/container"
	"github.com/dlomanov/mon/internal/apps/server/entrypoints/grpc/v1/interceptor"
	"github.com/dlomanov/mon/internal/apps/server/entrypoints/grpc/v1/services"
	pb "github.com/dlomanov/mon/internal/apps/shared/proto"
	"github.com/dlomanov/mon/internal/infra/grpcserver"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UseServices(s *grpcserver.Server, c *container.Container) {
	pb.RegisterMetricServiceServer(s.Server, services.NewMetricService(c.Logger, c.MetricUseCase))
}

func GetServerOptions(c *container.Container) grpcserver.Option {
	return grpcserver.ServerOptions(grpc.ChainUnaryInterceptor(
		interceptor.TrustedSubnet(c.Logger, c.Config.TrustedSubnet),
		logging.UnaryServerInterceptor(interceptorLogger(c.Logger.Sugar())),
		recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(func(p any) (err error) {
			c.Logger.Error("cached panic", zap.Any("panic", p))
			return status.Error(codes.Internal, "internal server error")
		}))))
}

func interceptorLogger(sugar *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			sugar.Debugf(msg, fields...)
		case logging.LevelInfo:
			sugar.Infof(msg, fields...)
		case logging.LevelWarn:
			sugar.Warnf(msg, fields...)
		case logging.LevelError:
			sugar.Errorf(msg, fields...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
