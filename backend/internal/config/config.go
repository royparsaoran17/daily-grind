package config

import (
	"bufio"
	"os"
	"strings"
)

// Config holds runtime configuration loaded from the environment.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	CORSOrigin  string
}

// Load reads configuration from a .env file (if present) and the environment.
// Real environment variables take precedence over .env values.
func Load() Config {
	loadDotEnv(".env")

	return Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://dailygrind:dailygrind@localhost:5432/dailygrind?sslmode=disable"),
		JWTSecret:   getenv("JWT_SECRET", "dev-secret-change-me-please-32chars"),
		CORSOrigin:  getenv("CORS_ORIGIN", "http://localhost:3000"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadDotEnv performs a minimal .env parse without pulling in a dependency.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}
