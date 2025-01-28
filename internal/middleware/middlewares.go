package middleware

import (
	"net/http"
	"strings"

	"github.com/runtime-metrics-course/internal/compress"
)

// func CompressMdlwr(next http.Handler) http.Handler {
// 	return http.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
// 		nextWriter := w

// 		encoding := r.Header.Get("Accept-Encoding")

// 		if strings.Contains(encoding, "gzip") {
// 			gw := compress.NewCompressedWriter(w)
// 			nextWriter = gw
// 			defer gw.Close()
// 		}

// 		contentEncoding := r.Header.Get("Content-Encoding")

// 		if strings.Contains(contentEncoding, "gzip") {

// 			gr, err := compress.NewCompressReader(r.Body)
// 			if err != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}
// 			r.Body = gr
// 			defer gr.Close()
// 		}

// 		next.ServeHTTP(nextWriter, r)

// 	})
// }
func CompressMdlwr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		if supportsGzip {
			cw := compress.NewCompressedWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		}

		next.ServeHTTP(ow, r)
	})
}
