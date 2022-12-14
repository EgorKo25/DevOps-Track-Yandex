package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"time"
)

type ConfigurationAgent struct {
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
	Address        string        `env:"ADDRESS"`
}

func NewAgentConfig() *ConfigurationAgent {
	var cfg ConfigurationAgent
	flag.StringVar(&cfg.Address,
		"a", "127.0.0.1:8080",
		"listening address of the server",
	)
	flag.DurationVar(&cfg.PollInterval,
		"p", time.Second*2,
		"timeout for update metrics",
	)
	flag.DurationVar(&cfg.ReportInterval,
		"r", time.Second*5,
		"timeout for report metrics",
	)
	flag.Parse()

	_ = env.Parse(&cfg)
	return &cfg
}

type ConfigurationServer struct {
	Address string `env:"ADDRESS" `

	StoreInterval time.Duration `env:"STORE_INTERVAL"`
	StoreFile     string        `env:"STORE_FILE"`
	Restore       bool          `env:"RESTORE"`
}

func NewServerConfig() *ConfigurationServer {
	var cfg ConfigurationServer
	flag.StringVar(&cfg.Address,
		"a", "127.0.0.1:8080",
		"listening address of the server",
	)
	flag.DurationVar(&cfg.StoreInterval,
		"i", time.Second*300,
		"timeout for save metrics",
	)
	flag.BoolVar(&cfg.Restore,
		"r", true,
		"recover files from storage",
	)
	flag.StringVar(&cfg.StoreFile,
		"f", "/tmp/devops-metrics-db.json",
		"file for saving metrics",
	)

	flag.Parse()

	_ = env.Parse(&cfg)

	return &cfg
}
