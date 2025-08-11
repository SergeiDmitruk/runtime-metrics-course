package interceptors

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/runtime-metrics-course/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type CryptoInterceptor struct {
	privateKey *rsa.PrivateKey
}

func NewCryptoInterceptor(privateKeyPath string) (*CryptoInterceptor, error) {
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBytes)
	if err != nil {
		logger.Log.Error(err.Error())
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &CryptoInterceptor{privateKey: privateKey}, nil
}

func (c *CryptoInterceptor) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "metadata is required")
		}

		if !c.requiresDecryption(md) {
			return handler(ctx, req)
		}

		reqProto, ok := req.(proto.Message)
		if !ok {
			return nil, status.Error(codes.Internal, "invalid request type")
		}

		encryptedData, err := proto.Marshal(reqProto)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to marshal request: %v", err)
		}

		decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, c.privateKey, encryptedData)
		if err != nil {
			logger.Log.Error(err.Error())
			return nil, status.Error(codes.InvalidArgument, "failed to decrypt data")
		}

		newReq := proto.Clone(reqProto)
		if err := proto.Unmarshal(decryptedData, newReq); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal decrypted data: %v", err)
		}

		return handler(ctx, newReq)
	}
}

func (c *CryptoInterceptor) requiresDecryption(md metadata.MD) bool {
	vals := md.Get("encrypted")
	return len(vals) > 0 && vals[0] == "true"
}
