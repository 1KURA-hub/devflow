package config

import (
	"strings"
	"testing"
)

func TestConfigValidateProductionJWTSecret(t *testing.T) {
	strongSecret := strings.Repeat("x", minProductionJWTSecretLength)
	tests := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{name: "empty", secret: "", wantErr: true},
		{name: "default", secret: defaultJWTSecret, wantErr: true},
		{name: "placeholder", secret: "change-me", wantErr: true},
		{name: "too short", secret: strings.Repeat("x", minProductionJWTSecretLength-1), wantErr: true},
		{name: "minimum accepted length", secret: strongSecret, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := (Config{AppEnv: " PROD ", JWTSecret: tt.secret}).Validate()
			if tt.wantErr && err == nil {
				t.Fatal("Validate() error = nil, want production JWT validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}
		})
	}
}

func TestConfigValidateAllowsDevelopmentDefault(t *testing.T) {
	err := (Config{AppEnv: "dev", JWTSecret: defaultJWTSecret}).Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v, want nil for development", err)
	}
}

func TestLoadPanicsForUnsafeProductionJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("JWT_SECRET", "")

	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("Load() did not panic for an empty production JWT_SECRET")
		}
	}()
	_ = Load()
}
