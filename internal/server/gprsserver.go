package server

import (
	"context"
	"fmt"
	"net"

	"github.com/runtime-metrics-course/internal/interceptors"
	"github.com/runtime-metrics-course/internal/logger"
	"github.com/runtime-metrics-course/internal/models"
	"github.com/runtime-metrics-course/internal/resilience"
	"github.com/runtime-metrics-course/internal/storage"
	"github.com/runtime-metrics-course/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	proto.UnimplementedMetricsServiceServer
	server  *grpc.Server
	storage storage.StorageIface
}

func NewGRPCServer(
	grpcAddress string,
	storage storage.StorageIface,
	secretKey string,
	subnet string,
	privateKeyPath string,
) (*GRPCServer, error) {

	opts := []grpc.ServerOption{}
	var intrs []grpc.UnaryServerInterceptor

	intrs = append(intrs, interceptors.GRPCLoggerInterceptor())

	if subnet != "" {
		whitelist := interceptors.NewGRPCWhiteListInterceptor(subnet)
		intrs = append(intrs, whitelist.UnaryInterceptor())
	}

	if secretKey != "" {
		hashInterceptor := interceptors.NewHashInterceptor([]byte(secretKey))
		intrs = append(intrs, hashInterceptor.UnaryInterceptor())
	}

	if privateKeyPath != "" {
		cryptoInterceptor, err := interceptors.NewCryptoInterceptor(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create crypto interceptor: %w", err)
		}
		intrs = append(intrs, cryptoInterceptor.UnaryInterceptor())
	}

	server := grpc.NewServer(opts...)

	return &GRPCServer{server: server, storage: storage}, nil
}

func (s *GRPCServer) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	logger.Log.Sugar().Infof("gRPC server starting on %s", address)
	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}
func (s *GRPCServer) UpdateMetric(ctx context.Context, req *proto.MetricRequest) (*proto.MetricResponse, error) {
	switch req.Mtype {
	case Gauge:
		val := req.GetGaugeValue()
		err := resilience.Retry(ctx, func() error {
			return s.storage.UpdateGauge(ctx, req.Id, val)
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &proto.MetricResponse{
			Id:    req.Id,
			Mtype: req.Mtype,
			Value: &proto.MetricResponse_GaugeValue{GaugeValue: val},
		}, nil

	case Counter:
		delta := req.GetCounterDelta()
		err := resilience.Retry(ctx, func() error {
			return s.storage.UpdateCounter(ctx, req.Id, delta)
		})
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		return &proto.MetricResponse{
			Id:    req.Id,
			Mtype: req.Mtype,
			Value: &proto.MetricResponse_CounterDelta{CounterDelta: delta},
		}, nil

	default:
		logger.Log.Error("Invalid metric type")
		return nil, status.Error(codes.InvalidArgument, "invalid metric type")
	}
}

func (s *GRPCServer) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.MetricResponse, error) {
	metrics, err := s.storage.GetMetrics(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	switch req.Mtype {
	case Gauge:
		if val, ok := metrics.Gauges[req.Id]; ok {
			return &proto.MetricResponse{
				Id:    req.Id,
				Mtype: req.Mtype,
				Value: &proto.MetricResponse_GaugeValue{GaugeValue: val},
			}, nil
		}
	case Counter:
		if val, ok := metrics.Counters[req.Id]; ok {
			return &proto.MetricResponse{
				Id:    req.Id,
				Mtype: req.Mtype,
				Value: &proto.MetricResponse_CounterDelta{CounterDelta: val},
			}, nil
		}
	default:
		logger.Log.Error("Invalid metric type")
		return nil, status.Error(codes.InvalidArgument, "invalid metric type")
	}

	return nil, status.Error(codes.NotFound, "metric not found")
}

func (s *GRPCServer) UpdateMetrics(ctx context.Context, req *proto.MetricsListRequest) (*proto.MetricsListResponse, error) {
	var metrics []models.MetricJSON

	for _, m := range req.Metrics {
		metric := models.MetricJSON{
			ID:    m.Id,
			MType: m.Mtype,
		}

		switch m.Mtype {
		case Gauge:
			if gv, ok := m.Value.(*proto.MetricRequest_GaugeValue); ok {
				val := gv.GaugeValue
				metric.Value = &val
			} else {
				logger.Log.Error("Invalid metric type")
				continue
			}
		case Counter:
			if cd, ok := m.Value.(*proto.MetricRequest_CounterDelta); ok {
				delta := cd.CounterDelta
				metric.Delta = &delta
			} else {
				logger.Log.Error("Invalid metric type")
				continue
			}
		default:
			logger.Log.Error("Invalid metric type")
			continue
		}

		metrics = append(metrics, metric)
	}

	err := resilience.Retry(ctx, func() error {
		return s.storage.UpdateAll(ctx, metrics)
	})
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.MetricsListResponse{}, nil
}

func (s *GRPCServer) Ping(ctx context.Context, _ *proto.PingRequest) (*proto.PingResponse, error) {
	err := s.storage.Ping(ctx)
	if err != nil {
		return &proto.PingResponse{Success: false}, nil
	}
	return &proto.PingResponse{Success: true}, nil
}

func (s *GRPCServer) GetMetrics(ctx context.Context, _ *proto.Empty) (*proto.MetricsListResponse, error) {
	storageMetrics, err := s.storage.GetMetrics(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &proto.MetricsListResponse{}

	for name, val := range storageMetrics.Gauges {
		resp.Metrics = append(resp.Metrics, &proto.MetricResponse{
			Id:    name,
			Mtype: Gauge,
			Value: &proto.MetricResponse_GaugeValue{GaugeValue: val},
		})
	}

	for name, val := range storageMetrics.Counters {
		resp.Metrics = append(resp.Metrics, &proto.MetricResponse{
			Id:    name,
			Mtype: Counter,
			Value: &proto.MetricResponse_CounterDelta{CounterDelta: val},
		})
	}

	return resp, nil
}
