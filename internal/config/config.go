package config

import (
	"blazeginx/internal/routing"
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

type Storage struct {
	DefaultExpiration time.Duration `yaml:"default_expiration_time" env-default:"30s"`
	CleanupInterval   time.Duration `yaml:"cleanup_interval" env-default:"30s"`
}

type RateLimit struct {
	Enabled    bool          `yaml:"enabled" env-default:"true"`
	MaxTokens  uint          `yaml:"max_tokens" env-default:"100"`
	RefillRate time.Duration `yaml:"refill_rate" env-default:"30s"`
	Storage    Storage
}

type Static struct {
	Enabled bool   `yaml:"enabled" env-default:"false"`
	Root    string `yaml:"root" env-default:"./web/dist"`
}

type Server struct {
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout" env-default:"5s"`
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout" env-default:"60s"`
}

type Upstream struct {
	MaxConnsPerHost       uint          `yaml:"max_conns_per_host" env-default:"64"`
	MaxIdleConnsPerHost   uint          `yaml:"max_idle_conns_per_host" env-default:"32"`
	IdleConnTimeout       time.Duration `yaml:"idle_conn_timeout" env-default:"60s"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout" env-default:"2s"`
	Services              []Service     `yaml:"services" env-required:"true"`
}

type Service struct {
	Name        string   `yaml:"name" env-required:"true"`
	Path        string   `yaml:"path" env-required:"true"`
	Urls        []string `yaml:"urls" env-required:"true"`
	HealthPath  string   `yaml:"health_path" env-default:"/healthz"`
	StripPrefix bool     `yaml:"strip_prefix" env-default:"false"`
}

type Config struct {
	Env       Env    `yaml:"env" env-default:"local"`
	Addr      string `yaml:"addr" env-default:"127.0.0.1:8888"`
	AdminAddr string `yaml:"admin_addr" env-default:"127.0.0.1:9999"`

	Server    Server    `yaml:"server"`
	Upstream  Upstream  `yaml:"upstream" env-required:"true"`
	RateLimit RateLimit `yaml:"rate-limit"`
	Static    Static    `yaml:"static"`
}

func normalizeServices(c *Config) {
	wasPath := make(map[string]string)
	wasUrl := make(map[string]int)

	for _, s := range c.Upstream.Services {
		nPath, err := routing.NormalizeRoutePath(s.Path)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
		ent, ok := wasPath[nPath]
		if ok {
			log.Fatalf("Route paths for services: %s, %s cannot be same",
				ent, s.Name,
			)
		}
		wasPath[nPath] = s.Name
		s.Path = nPath

		for i, u := range s.Urls {
			nUrl, err := routing.NormalizeUrl(u)
			if err != nil {
				log.Fatalf("service %s: %s", s.Name, err.Error())
			}
			_, ok = wasUrl[nUrl]
			if ok {
				log.Fatalf("URL %s is used multiple times in services", nUrl)
			}
			wasUrl[nUrl] = 67
			s.Urls[i] = nUrl
		}
	}
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

	normalizeServices(&cfg)

	if !cfg.Env.IsValid() {
		log.Fatalf("Invalid enviroment type in config: %s", cfg.Env)
	}

	if cfg.Addr == cfg.AdminAddr {
		log.Fatalf("addr and admin_addr cannot be same")
	}
	return cfg
}
