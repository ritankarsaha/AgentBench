package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	AppEnv      string
	FrontendURL string

	DatabaseURL string

	SupabaseURL            string
	SupabaseServiceRoleKey string
	SupabaseStorageBucket  string

	NvidiaNIMAPIKeys      string
	NvidiaNIMBaseURL      string
	NvidiaNIMDefaultModel string

	AgentThreadsAPIURL         string
	AgentThreadsPlatformAPIKey string

	AgentReplayAPIURL       string
	AgentReplaySharedSecret string

	SandboxTimeoutSeconds int
	SandboxMemoryLimitMB  int
}

func Load() *Config {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "development"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		DatabaseURL: mustGetEnv("DATABASE_URL"),

		SupabaseURL:            mustGetEnv("SUPABASE_URL"),
		SupabaseServiceRoleKey: mustGetEnv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseStorageBucket:  getEnv("SUPABASE_STORAGE_BUCKET", "agentbench-traces"),

		NvidiaNIMAPIKeys:      getEnv("NVIDIA_NIM_API_KEYS", ""),
		NvidiaNIMBaseURL:      getEnv("NVIDIA_NIM_BASE_URL", "https://integrate.api.nvidia.com/v1"),
		NvidiaNIMDefaultModel: getEnv("NVIDIA_NIM_DEFAULT_MODEL", "meta/llama-3.1-70b-instruct"),

		AgentThreadsAPIURL:         getEnv("AGENTTHREADS_API_URL", ""),
		AgentThreadsPlatformAPIKey: getEnv("AGENTTHREADS_PLATFORM_API_KEY", ""),

		AgentReplayAPIURL:       getEnv("AGENTREPLAY_API_URL", ""),
		AgentReplaySharedSecret: getEnv("AGENTREPLAY_SHARED_SECRET", ""),

		SandboxTimeoutSeconds: getEnvInt("SANDBOX_TIMEOUT_SECONDS", 30),
		SandboxMemoryLimitMB:  getEnvInt("SANDBOX_MEMORY_LIMIT_MB", 256),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		panic(fmt.Sprintf("config: required env var %s is not set", key))
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	var i int
	if _, err := fmt.Sscanf(v, "%d", &i); err != nil {
		return fallback
	}
	return i
}
