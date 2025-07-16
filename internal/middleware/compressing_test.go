package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/runtime-metrics-course/internal/compress"
)

func TestCompressMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	ts := httptest.NewServer(CompressMiddleware(handler))
	defer ts.Close()

	client := &http.Client{}

	tests := []struct {
		name             string
		acceptEncoding   string
		contentEncoding  string
		expectedEncoding string
		expectedBody     string
		wantError        bool
		body             string
	}{
		{
			name:             "No Accept-Encoding: gzip",
			acceptEncoding:   "",
			expectedEncoding: "",
			expectedBody:     `{"status":"ok"}`,
			body:             `{"status":"ok"}`,
		},
		{
			name:             "Accept-Encoding: gzip, response should be gzipped",
			acceptEncoding:   "gzip",
			expectedEncoding: "gzip",
			expectedBody:     `{"status":"ok"}`,
			body:             `{"status":"ok"}`,
		},
		{
			name:             "Content-Encoding: gzip, request should be decompressed",
			contentEncoding:  "gzip",
			expectedEncoding: "",
			expectedBody:     `{"status":"ok"}`,
			body:             `{"status":"ok"}`,
		},
		{
			name:             "Error in gzip decompression",
			contentEncoding:  "gzip",
			acceptEncoding:   "gzip",
			expectedEncoding: "",
			wantError:        true,
			body:             `invalid body`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var requestBody io.Reader
			if tt.contentEncoding == "gzip" && !tt.wantError {
				gzippedBody, err := compress.CompressGzip([]byte(tt.body))
				if err != nil {
					t.Fatalf("gzip compression error: %v", err)
				}
				requestBody = bytes.NewReader(gzippedBody)
			} else {
				requestBody = strings.NewReader(tt.body)
			}

			req, _ := http.NewRequest(http.MethodPost, ts.URL, requestBody)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.Header.Get("Content-Encoding") != tt.expectedEncoding {
				t.Errorf("expected Content-Encoding: %s, got: %s", tt.expectedEncoding, resp.Header.Get("Content-Encoding"))
			}

			if tt.wantError && resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected status %d, got: %d", http.StatusBadRequest, resp.StatusCode)
			} else if !tt.wantError {
				var body []byte
				if tt.expectedEncoding == "gzip" {
					gr, err := gzip.NewReader(resp.Body)
					if err != nil {
						t.Fatalf("gzip reader error: %v", err)
					}
					body, _ = io.ReadAll(gr)
					gr.Close()
				} else {
					body, _ = io.ReadAll(resp.Body)
				}

				if string(body) != tt.expectedBody {
					t.Errorf("unexpected body: %s", body)
				}
			}
		})
	}
}

func BenchmarkCompressMiddleware(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	ts := httptest.NewServer(CompressMiddleware(handler))
	defer ts.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("Accept-Encoding", "gzip")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Do(req)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
	}
}
