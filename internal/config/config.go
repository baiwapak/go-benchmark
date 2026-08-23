package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/yourorg/go-benchmark/internal/capacity"
)

type Config struct {
	ServerAddr            string
	ServerMode            string
	DefaultLang           string
	ReportDir             string
	MaxConcurrency        int
	MaxDurationSec        int
	MaxTimeoutSec         int
	MaxInflightJobs       int
	MaxRampSec            int
	MaxBodyBytes          int
	CreateRateLimitPerMin int
	GitHubRepoURL         string
	Host                  capacity.Host
	LimitsAuto            bool
	CORSAllowOrigins      []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	host := capacity.DetectHost()
	rec := capacity.Recommend(host)
	limitsAuto := !envSet("MAX_CONCURRENCY") && !envSet("MAX_DURATION_SEC") && !envSet("MAX_INFLIGHT_JOBS")

	cfg := &Config{
		ServerAddr:            getenv("SERVER_ADDR", ":8000"),
		ServerMode:            getenv("SERVER_MODE", "debug"),
		DefaultLang:           getenv("DEFAULT_LANG", "zh"),
		ReportDir:             getenv("REPORT_DIR", "data/reports"),
		MaxConcurrency:        getenvInt("MAX_CONCURRENCY", rec.MaxConcurrency),
		MaxDurationSec:        getenvInt("MAX_DURATION_SEC", rec.MaxDurationSec),
		MaxTimeoutSec:         getenvInt("MAX_TIMEOUT_SEC", 60),
		MaxInflightJobs:       getenvInt("MAX_INFLIGHT_JOBS", rec.MaxInflightJobs),
		MaxRampSec:            getenvInt("MAX_RAMP_SEC", 120),
		MaxBodyBytes:          getenvInt("MAX_BODY_BYTES", 65536),
		CreateRateLimitPerMin: getenvInt("CREATE_RATE_LIMIT_PER_MIN", 10),
		GitHubRepoURL:         getenv("GITHUB_REPO_URL", "https://github.com/yourorg/go-benchmark"),
		Host:                  host,
		LimitsAuto:            limitsAuto,
	}

	if origins := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")); origins != "" {
		for _, o := range strings.Split(origins, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				cfg.CORSAllowOrigins = append(cfg.CORSAllowOrigins, o)
			}
		}
	} else if cfg.ServerMode == "debug" {
		cfg.CORSAllowOrigins = []string{"*"}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DefaultLang != "zh" && c.DefaultLang != "en" {
		return fmt.Errorf("DEFAULT_LANG must be zh or en")
	}
	if c.MaxConcurrency < 1 {
		return fmt.Errorf("MAX_CONCURRENCY must be >= 1")
	}
	if c.MaxDurationSec < 1 {
		return fmt.Errorf("MAX_DURATION_SEC must be >= 1")
	}
	if c.MaxTimeoutSec < 1 {
		return fmt.Errorf("MAX_TIMEOUT_SEC must be >= 1")
	}
	if c.MaxInflightJobs < 1 {
		return fmt.Errorf("MAX_INFLIGHT_JOBS must be >= 1")
	}
	if c.ReportDir == "" {
		return fmt.Errorf("REPORT_DIR is required")
	}
	return nil
}

func envSet(key string) bool {
	v, ok := os.LookupEnv(key)
	return ok && strings.TrimSpace(v) != ""
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
