package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

type respData struct {
	statusCode int
	size       int
}
type loggerResponseWriter struct {
	http.ResponseWriter
	respData *respData
}

func (r *loggerResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.respData.statusCode = statusCode
}
func (r *loggerResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.respData.size += size
	return size, err
}
func Init(level string) error {
	l, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = l
	logger, err := cfg.Build()
	Log = logger
	if err != nil {
		return err
	}
	return nil
}

func LoggerMdlwr(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggerResponseWriter{
			ResponseWriter: w,
			respData:       &respData{},
		}
		h.ServeHTTP(&lw, r)

		Log.Sugar().Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", lw.respData.statusCode,
			"size", lw.respData.size,
			"duration", time.Since(start),
		)
	}
}
