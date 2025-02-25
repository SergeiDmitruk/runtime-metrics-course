package middleware

import (
	"net/http"
	"strings"

	"github.com/runtime-metrics-course/internal/utils"
)

func CompressMdlwr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		cw := utils.NewCompressedWriter(w)
		if supportsGzip {
			w = cw
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		if contentEncoding == "gzip" {
			cr, err := utils.NewCompressReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decompress gzip body", http.StatusBadRequest)
				return
			}
			defer cr.Close()
			r.Body = cr
		}

		next.ServeHTTP(w, r)

		if cw != nil && cw.NeedCompress {
			cw.Close()
		}
	})
}
