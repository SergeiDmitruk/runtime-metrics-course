package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendRequest(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %v", r.Method)
		}
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("expected Content-Type text/plain, got %v", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tests := []struct {
		name      string
		url       string
		client    *http.Client
		expectErr bool
	}{
		{
			name:      "Successful request",
			url:       ts.URL,
			client:    &http.Client{},
			expectErr: false,
		},
		{
			name:      "Invalid URL",
			url:       "invalid-url",
			client:    &http.Client{},
			expectErr: true,
		},
		{
			name:      "Request fails",
			url:       "http://nonexistent-url",
			client:    &http.Client{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sendRequest(tt.client, tt.url)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err != nil)
			}
		})
	}
}
