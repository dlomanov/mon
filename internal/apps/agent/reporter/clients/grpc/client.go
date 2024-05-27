package grpc

import (
	"context"
	"github.com/dlomanov/mon/internal/apps/agent/reporter"
	"github.com/dlomanov/mon/internal/apps/agent/reporter/utils"
	pb "github.com/dlomanov/mon/internal/apps/shared/proto"
	"github.com/dlomanov/mon/internal/entities"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var _ reporter.Client = (*Client)(nil)

type (
	Client struct {
		logger *zap.Logger
		conn   *grpc.ClientConn
		client pb.MetricServiceClient
	}
)

func New(
	logger *zap.Logger,
	grpcAddr string,
) (*Client, error) {
	conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := pb.NewMetricServiceClient(conn)

	return &Client{
		logger: logger,
		conn:   conn,
		client: client,
	}, nil
}

func (r *Client) Report(ctx context.Context, metrics map[string]entities.Metric) {
	ip, err := utils.GetOutboundIP()
	if err != nil {
		r.logger.Error("get outbound ip failed", zap.Error(err))
	} else {
		ctx = metadata.AppendToOutgoingContext(ctx, "X-Real-IP", ip.String())
	}

	if _, err := r.client.Update(ctx, &pb.UpdateRequest{Metrics: r.toModels(metrics)}); err != nil {
		r.logger.Error("failed report metrics", zap.Error(err))
	}
}

func (r *Client) Close() error {
	return r.conn.Close()
}

func (r *Client) toModels(metrics map[string]entities.Metric) []*pb.Metric {
	mapType := func(t entities.MetricType) pb.MetricType {
		switch t {
		case entities.MetricCounter:
			return pb.MetricType_COUNTER
		case entities.MetricGauge:
			return pb.MetricType_GAUGE
		default:
			return pb.MetricType_UNKNOWN
		}
	}

	ms := make([]*pb.Metric, 0, len(metrics))
	for _, v := range metrics {
		ms = append(ms, &pb.Metric{
			Name:  v.Name,
			Type:  mapType(v.Type),
			Value: v.Value,
			Delta: v.Delta,
		})
	}
	return ms
}
