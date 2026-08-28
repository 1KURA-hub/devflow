package main

import "testing"

func TestEnsureSeedAllowed(t *testing.T) {
	tests := []struct {
		name    string
		appEnv  string
		wantErr bool
	}{
		{name: "production", appEnv: "prod", wantErr: true},
		{name: "production is case insensitive", appEnv: " PROD ", wantErr: true},
		{name: "development", appEnv: "dev", wantErr: false},
		{name: "test", appEnv: "test", wantErr: false},
		{name: "empty", appEnv: "", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureSeedAllowed(tt.appEnv)
			if tt.wantErr && err == nil {
				t.Fatal("ensureSeedAllowed() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ensureSeedAllowed() error = %v, want nil", err)
			}
		})
	}
}
