package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

type HashMiddleware struct {
	key []byte
}

func NewHashMiddleware(key []byte) *HashMiddleware {
	return &HashMiddleware{key: key}
}
func (h *HashMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			r.Body.Close()
			r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(body)))

			clientHash := r.Header.Get("HashSHA256")
			expectedHash := HmacSHA256(body, h.key)

			if clientHash != "" && clientHash != expectedHash {
				http.Error(w, "Invalid Hash", http.StatusBadRequest)
				return
			}
		}

		rec := &responseWriterInterceptor{ResponseWriter: w, buffer: new(bytes.Buffer)}
		next.ServeHTTP(rec, r)

		responseHash := HmacSHA256(rec.buffer.Bytes(), h.key)
		w.Header().Set("HashSHA256", responseHash)

		w.Write(rec.buffer.Bytes())
	})
}

type responseWriterInterceptor struct {
	http.ResponseWriter
	buffer *bytes.Buffer
}

func (rw *responseWriterInterceptor) Write(data []byte) (int, error) {
	return rw.buffer.Write(data)
}
func HmacSHA256(data []byte, key []byte) string {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
