// crypto_middleware.go
package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/runtime-metrics-course/internal/logger"
)

type CryptoMiddleware struct {
	privateKey *rsa.PrivateKey
}

func NewCryptoMiddleware(privateKeyPath string) (*CryptoMiddleware, error) {
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

	return &CryptoMiddleware{privateKey: privateKey}, nil
}

func (m *CryptoMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encryptedData, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, "Failed to read encrypted body", http.StatusBadRequest)
			return
		}

		decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, m.privateKey, encryptedData)
		if err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, "Failed to decrypt data", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(decryptedData))
		r.ContentLength = int64(len(decryptedData))

		next.ServeHTTP(w, r)
	})
}
