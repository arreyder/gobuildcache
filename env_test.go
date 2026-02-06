package main

import (
	"os"
	"testing"
)

func TestGetEnvWithPrefix(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envVars      map[string]string
		expected     string
	}{
		{
			name:         "returns default when neither env var is set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envVars:      map[string]string{},
			expected:     "default",
		},
		{
			name:         "returns unprefixed value when only unprefixed is set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envVars:      map[string]string{"TEST_KEY": "unprefixed_value"},
			expected:     "unprefixed_value",
		},
		{
			name:         "returns prefixed value when only prefixed is set",
			key:          "TEST_KEY",
			defaultValue: "default",
			envVars:      map[string]string{"GOBUILDCACHE_TEST_KEY": "prefixed_value"},
			expected:     "prefixed_value",
		},
		{
			name:         "prefixed value takes precedence over unprefixed",
			key:          "TEST_KEY",
			defaultValue: "default",
			envVars: map[string]string{
				"TEST_KEY":             "unprefixed_value",
				"GOBUILDCACHE_TEST_KEY": "prefixed_value",
			},
			expected: "prefixed_value",
		},
		{
			name:         "works with S3_BUCKET style keys",
			key:          "S3_BUCKET",
			defaultValue: "",
			envVars:      map[string]string{"GOBUILDCACHE_S3_BUCKET": "my-bucket"},
			expected:     "my-bucket",
		},
		{
			name:         "works with BACKEND_TYPE style keys",
			key:          "BACKEND_TYPE",
			defaultValue: "disk",
			envVars:      map[string]string{"GOBUILDCACHE_BACKEND_TYPE": "s3"},
			expected:     "s3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set environment variables
			os.Unsetenv(tt.key)
			os.Unsetenv("GOBUILDCACHE_" + tt.key)
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := getEnvWithPrefix(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvWithPrefix(%q, %q) = %q, want %q", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvBoolWithPrefix(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envVars      map[string]string
		expected     bool
	}{
		{
			name:         "returns default when neither env var is set",
			key:          "TEST_BOOL",
			defaultValue: false,
			envVars:      map[string]string{},
			expected:     false,
		},
		{
			name:         "returns true default when neither env var is set",
			key:          "TEST_BOOL",
			defaultValue: true,
			envVars:      map[string]string{},
			expected:     true,
		},
		{
			name:         "returns unprefixed value when only unprefixed is set",
			key:          "TEST_BOOL",
			defaultValue: false,
			envVars:      map[string]string{"TEST_BOOL": "true"},
			expected:     true,
		},
		{
			name:         "returns prefixed value when only prefixed is set",
			key:          "TEST_BOOL",
			defaultValue: false,
			envVars:      map[string]string{"GOBUILDCACHE_TEST_BOOL": "true"},
			expected:     true,
		},
		{
			name:         "prefixed value takes precedence over unprefixed",
			key:          "TEST_BOOL",
			defaultValue: false,
			envVars: map[string]string{
				"TEST_BOOL":             "false",
				"GOBUILDCACHE_TEST_BOOL": "true",
			},
			expected: true,
		},
		{
			name:         "prefixed false overrides unprefixed true",
			key:          "TEST_BOOL",
			defaultValue: true,
			envVars: map[string]string{
				"TEST_BOOL":             "true",
				"GOBUILDCACHE_TEST_BOOL": "false",
			},
			expected: false,
		},
		{
			name:         "accepts 1 as true",
			key:          "TEST_BOOL",
			defaultValue: false,
			envVars:      map[string]string{"GOBUILDCACHE_TEST_BOOL": "1"},
			expected:     true,
		},
		{
			name:         "accepts yes as true",
			key:          "TEST_BOOL",
			defaultValue: false,
			envVars:      map[string]string{"GOBUILDCACHE_TEST_BOOL": "yes"},
			expected:     true,
		},
		{
			name:         "accepts YES as true (case insensitive)",
			key:          "TEST_BOOL",
			defaultValue: false,
			envVars:      map[string]string{"GOBUILDCACHE_TEST_BOOL": "YES"},
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set environment variables
			os.Unsetenv(tt.key)
			os.Unsetenv("GOBUILDCACHE_" + tt.key)
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := getEnvBoolWithPrefix(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvBoolWithPrefix(%q, %v) = %v, want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetEnvFloatWithPrefix(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue float64
		envVars      map[string]string
		expected     float64
	}{
		{
			name:         "returns default when neither env var is set",
			key:          "TEST_FLOAT",
			defaultValue: 0.5,
			envVars:      map[string]string{},
			expected:     0.5,
		},
		{
			name:         "returns unprefixed value when only unprefixed is set",
			key:          "TEST_FLOAT",
			defaultValue: 0.0,
			envVars:      map[string]string{"TEST_FLOAT": "0.75"},
			expected:     0.75,
		},
		{
			name:         "returns prefixed value when only prefixed is set",
			key:          "TEST_FLOAT",
			defaultValue: 0.0,
			envVars:      map[string]string{"GOBUILDCACHE_TEST_FLOAT": "0.25"},
			expected:     0.25,
		},
		{
			name:         "prefixed value takes precedence over unprefixed",
			key:          "TEST_FLOAT",
			defaultValue: 0.0,
			envVars: map[string]string{
				"TEST_FLOAT":             "0.5",
				"GOBUILDCACHE_TEST_FLOAT": "0.9",
			},
			expected: 0.9,
		},
		{
			name:         "returns default for invalid prefixed value, falls back to unprefixed",
			key:          "TEST_FLOAT",
			defaultValue: 0.0,
			envVars: map[string]string{
				"TEST_FLOAT":             "0.5",
				"GOBUILDCACHE_TEST_FLOAT": "not-a-number",
			},
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set environment variables
			os.Unsetenv(tt.key)
			os.Unsetenv("GOBUILDCACHE_" + tt.key)
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			result := getEnvFloatWithPrefix(tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvFloatWithPrefix(%q, %v) = %v, want %v", tt.key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
