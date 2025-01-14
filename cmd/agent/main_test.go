package main

import (
	"errors"
	"testing"
)

func TestSet(t *testing.T) {
	tests := []struct {
		input        string
		expectedHost string
		expectedPort int
		expectedErr  error
	}{
		{
			input:        "localhost:8080",
			expectedHost: "localhost",
			expectedPort: 8080,
			expectedErr:  nil,
		},
		{
			input:        "127.0.0.1:9090",
			expectedHost: "127.0.0.1",
			expectedPort: 9090,
			expectedErr:  nil,
		},
		{
			input:        "invalid-format",
			expectedHost: "",
			expectedPort: 0,
			expectedErr:  errors.New("need Config in a form host:port"),
		},
		{
			input:        "localhost:abc",
			expectedHost: "",
			expectedPort: 0,
			expectedErr:  errors.New("strconv.Atoi: parsing \"abc\": invalid syntax"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			config := &Config{}
			err := config.Set(tt.input)

			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}

			if config.Host != tt.expectedHost {
				t.Errorf("expected Host %v, got %v", tt.expectedHost, config.Host)
			}
			if config.Port != tt.expectedPort {
				t.Errorf("expected Port %v, got %v", tt.expectedPort, config.Port)
			}
		})
	}
}
