package interceptors

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type GRPCWhiteListInterceptor struct {
	subnet *net.IPNet
}

func NewGRPCWhiteListInterceptor(cidr string) *GRPCWhiteListInterceptor {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return &GRPCWhiteListInterceptor{subnet: nil}
	}
	return &GRPCWhiteListInterceptor{subnet: subnet}
}

func (w *GRPCWhiteListInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if w.subnet == nil {
			return handler(ctx, req)
		}

		var clientIP net.IP
		if p, ok := peer.FromContext(ctx); ok {
			switch addr := p.Addr.(type) {
			case *net.TCPAddr:
				clientIP = addr.IP
			case *net.UDPAddr:
				clientIP = addr.IP
			}
		}

		if clientIP == nil {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				if realIPs := md.Get("x-real-ip"); len(realIPs) > 0 {
					clientIP = net.ParseIP(realIPs[0])
				}
			}
		}

		if clientIP == nil {
			return nil, status.Errorf(codes.PermissionDenied, "client IP not detected")
		}

		if !w.subnet.Contains(clientIP) {
			return nil, status.Errorf(codes.PermissionDenied, "IP %s not in whitelist", clientIP.String())
		}

		return handler(ctx, req)
	}
}
