package config

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server: ServerConfig{
					Port:         8080,
					Host:         "0.0.0.0",
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{
					Level:  "info",
					Format: "json",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port - too low",
			config: Config{
				Server: ServerConfig{
					Port:         0,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{Level: "info"},
			},
			wantErr: true,
		},
		{
			name: "invalid port - too high",
			config: Config{
				Server: ServerConfig{
					Port:         70000,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{Level: "info"},
			},
			wantErr: true,
		},
		{
			name: "invalid read timeout",
			config: Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  0,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{Level: "info"},
			},
			wantErr: true,
		},
		{
			name: "empty log level",
			config: Config{
				Server: ServerConfig{
					Port:         8080,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
				},
				Log: LogConfig{Level: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServerConfig_GetAddress(t *testing.T) {
	cfg := ServerConfig{
		Host: "localhost",
		Port: 8080,
	}

	expected := "localhost:8080"
	if got := cfg.GetAddress(); got != expected {
		t.Errorf("GetAddress() = %v, want %v", got, expected)
	}
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	if got := cfg.GetDSN(); got != expected {
		t.Errorf("GetDSN() = %v, want %v", got, expected)
	}
}

func TestCacheConfig_GetRedisAddress(t *testing.T) {
	cfg := CacheConfig{
		Host: "redis",
		Port: 6379,
	}

	expected := "redis:6379"
	if got := cfg.GetRedisAddress(); got != expected {
		t.Errorf("GetRedisAddress() = %v, want %v", got, expected)
	}
}
