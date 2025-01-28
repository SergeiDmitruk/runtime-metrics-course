package middleware

import (
	"net/http"
	"strings"

	"github.com/runtime-metrics-course/internal/compress"
)

func CompressMdlwr(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nextWriter := w

		encoding := r.Header.Get("Accept-Encoding")

		if strings.Contains(encoding, "gzip") {
			gw := compress.NewCompressedWriter(w)
			nextWriter = gw
			defer gw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")

		if strings.Contains(contentEncoding, "gzip") {

			gr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = gr
			defer gr.Close()
		}

		next(nextWriter, r)

	}
}
