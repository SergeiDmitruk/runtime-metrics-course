package storage

import "testing"

func TestInitStorage(t *testing.T) {

	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{name: "Correct storage type",
			arg:     RuntimeMemory,
			wantErr: false},
		{name: "Incorrect storage type",
			arg:     "Unnknown type",
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := InitStorage(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("InitStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
