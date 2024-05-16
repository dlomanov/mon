package services

import (
	"context"
	"github.com/dlomanov/mon/internal/apps/server/usecases"
	pb "github.com/dlomanov/mon/internal/apps/shared/proto"
	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ pb.MetricServiceServer = (*MetricService)(nil)

type MetricService struct {
	pb.UnimplementedMetricServiceServer
	logger   *zap.Logger
	metricUC *usecases.MetricUseCase
}

func NewMetricService(
	logger *zap.Logger,
	metricUC *usecases.MetricUseCase,
) *MetricService {
	return &MetricService{
		logger:   logger,
		metricUC: metricUC,
	}
}

func (m *MetricService) Update(ctx context.Context, request *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	emptyResp := &pb.UpdateResponse{}

	metrics := request.GetMetrics()
	if (metrics == nil) || (len(metrics) == 0) {
		m.logger.Debug("no metrics provided")
		return emptyResp, status.Error(codes.InvalidArgument, "no metrics provided")
	}
	ms, err := m.toEntities(metrics)
	if err != nil {
		m.logger.Debug("failed map to entities", zap.Error(err))
		return emptyResp, status.Error(codes.InvalidArgument, err.Error())
	}

	if _, err := m.metricUC.Update(ctx, ms...); err != nil {
		m.logger.Debug("failed update metrics", zap.Error(err))
		return emptyResp, status.Error(codes.Internal, err.Error())
	}

	return emptyResp, nil
}

func (m *MetricService) toEntities(metrics []*pb.Metric) ([]entities.Metric, error) {
	var (
		mapType = func(t pb.MetricType) (entities.MetricType, error) {
			switch t {
			case pb.MetricType_COUNTER:
				return entities.MetricCounter, nil
			case pb.MetricType_GAUGE:
				return entities.MetricGauge, nil
			default:
				return entities.MetricUnknown, status.Error(codes.InvalidArgument, "unknown metric type")
			}
		}
		mapMetric = func(m *pb.Metric) (entities.Metric, error) {
			entity := entities.Metric{}
			typ, err := mapType(m.Type)
			switch {
			case err != nil:
				return entity, err
			case typ == entities.MetricCounter && (m.Delta == nil || m.Value != nil):
				return entity, status.Error(codes.InvalidArgument, "invalid metric type")
			case typ == entities.MetricGauge && (m.Delta != nil || m.Value == nil):
				return entity, status.Error(codes.InvalidArgument, "invalid metric type")
			default:
				entity.Name = m.Name
				entity.Type = typ
				entity.Delta = m.Delta
				entity.Value = m.Value
				return entity, nil
			}
		}
	)

	result := make([]entities.Metric, 0, len(metrics))
	for _, metric := range metrics {
		entity, err := mapMetric(metric)
		if err != nil {
			return nil, err
		}
		result = append(result, entity)
	}
	return result, nil
}
