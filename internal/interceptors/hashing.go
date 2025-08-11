package interceptors

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type HashInterceptor struct {
	key []byte
}

func NewHashInterceptor(key []byte) *HashInterceptor {
	return &HashInterceptor{key: key}
}

func (h *HashInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if h.isReadOnlyMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is required for authentication")
		}
		clientHash := h.getFirstMetadataValue(md, "hashsha256")
		if clientHash == "" {
			return nil, status.Error(codes.Unauthenticated, "hashsha256 header is required")
		}

		reqProto, ok := req.(proto.Message)
		if !ok {
			return nil, status.Error(codes.Internal, "invalid request type, expected protobuf message")
		}

		reqBytes, err := proto.Marshal(reqProto)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to marshal request: %v", err)
		}

		if !h.verifyHash(reqBytes, clientHash) {
			return nil, status.Error(codes.PermissionDenied, "invalid request hash")
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		respProto, ok := resp.(proto.Message)
		if !ok {
			return nil, status.Error(codes.Internal, "invalid response type, expected protobuf message")
		}

		respBytes, err := proto.Marshal(respProto)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal response: %v", err)
		}

		responseHash := h.generateHash(respBytes)

		err = grpc.SetTrailer(ctx, metadata.Pairs("hashsha256", responseHash))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to set response hash: %v", err)
		}

		return resp, nil
	}
}

func (h *HashInterceptor) isReadOnlyMethod(fullMethod string) bool {
	readOnlyMethods := map[string]bool{
		"/metrics.MetricsService/GetMetric":  true,
		"/metrics.MetricsService/GetMetrics": true,
		"/metrics.MetricsService/Ping":       true,
	}

	return readOnlyMethods[fullMethod]
}

func (h *HashInterceptor) getFirstMetadataValue(md metadata.MD, key string) string {
	values := md.Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (h *HashInterceptor) verifyHash(data []byte, receivedHash string) bool {
	expectedHash := h.generateHash(data)
	return hmac.Equal([]byte(expectedHash), []byte(receivedHash))
}

func (h *HashInterceptor) generateHash(data []byte) string {
	mac := hmac.New(sha256.New, h.key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}
