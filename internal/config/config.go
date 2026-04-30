package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Env string

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

func (e Env) IsValid() bool {
	switch e {
	case EnvLocal, EnvDev, EnvProd:
		return true
	default:
		return false
	}
}

type RateLimit struct {
	Enabled           bool          `yaml:"enabled" env-default:"true"`
	MaxTokens         uint          `yaml:"max_tokens" env-default:"100"`
	RefillRate        time.Duration `yaml:"refill_rate" env-default:"30s"`
	DefaultExpiration time.Duration `yaml:"default_expiration_time" env-default:"30s"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval" env-default:"30s"`
}

type Timeout struct {
	Upstream time.Duration `yaml:"upstream" env-default:"2s"`
	Server   time.Duration `yaml:"server" env-default:"5s"`
	Idle     time.Duration `yaml:"idle" env-default:"60s"`
}

type Static struct {
	Enabled bool   `yaml:"enabled" env-default:"false"`
	Root    string `yaml:"root" env-default:"./web/dist"`
}

type Route struct {
	Path        string `yaml:"path" env-required:"true"`
	Url         string `yaml:"url" env-required:"true"`
	StripPrefix bool   `yaml:"strip_prefix" env-default:"false"`
}

type Config struct {
	Env  Env    `yaml:"env" env-default:"local"`
	Addr string `yaml:"addr" env-default:"127.0.0.1:8888"`

	Routes    []Route   `yaml:"routes" env-required:"true"`
	RateLimit RateLimit `yaml:"rate-limit"`
	Timeout   Timeout   `yaml:"timeout"`
	Static    Static    `yaml:"static"`
}

func validateRoutes(c *Config) {
	for _, s := range c.Routes {
		if s.Path == "" {
			log.Fatalf("Route cannot be empty")
		}
	}
}

func validateEnv(c *Config) {
	if !c.Env.IsValid() {
		log.Fatalf("Invalid enviroment type in config: %s", c.Env)
	}
}

func validate(c *Config) {
	validateRoutes(c)
	validateEnv(c)
}

func MustRead() Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("Config is missing, please set CONFIG_PATH enviroment variable\n")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file not found: %s", configPath)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Invalid config: %s", err.Error())
	}

	validate(&cfg)
	return cfg
}
