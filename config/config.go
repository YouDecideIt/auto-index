package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type NgMonitorConfig struct {
	Address string `yaml:"address"`
}

type TiDBConfig struct {
	Address string `yaml:"address"`
}

type WebConfig struct {
	Address string `yaml:"address"`
}

type StaticConfig struct {
	Targets []string `yaml:"targets"`
}

type ScrapeConfig struct {
	ScrapeInterval time.Duration `yaml:"scrape_interval"`
	ScrapeTimeout  time.Duration `yaml:"scrape_timeout"`
}

type LogConfig struct {
	LogLevel string `yaml:"log_level"`
	LogFile  string `yaml:"log_file"`
}

type EvaluateConfig struct {
	Interval time.Duration `yaml:"interval"`
}

type AutoIndexConfig struct {
	TiDBConfig      TiDBConfig      `yaml:"tidb"`
	NgMonitorConfig NgMonitorConfig `yaml:"ng-monitor"`
	WebConfig       WebConfig       `yaml:"web"`
	ScrapeConfigs   []*ScrapeConfig `yaml:"scrape_configs"`
	LogConfig       LogConfig       `yaml:"logs"`
	EvaluateConfig  EvaluateConfig  `yaml:"evaluate"`
}

var DefaultAutoIndexConfig = AutoIndexConfig{
	TiDBConfig: TiDBConfig{
		Address: "127.0.0.1:4000",
	},
	WebConfig: WebConfig{
		Address: "127.0.0.1:9977",
	},
	LogConfig: LogConfig{
		LogLevel: "debug",
	},
	EvaluateConfig: EvaluateConfig{
		Interval: 0 * time.Second,
	},
}

func LoadConfig(cfgFilePath string, override func(config *AutoIndexConfig)) (*AutoIndexConfig, error) {
	cfg := DefaultAutoIndexConfig

	if cfgFilePath != "" {
		file, err := os.Open(cfgFilePath)
		if err != nil {
			return nil, err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				return
			}
		}(file)

		// Init new YAML decode
		d := yaml.NewDecoder(file)

		// Start YAML decoding from file
		if err = d.Decode(&cfg); err != nil {
			return nil, err
		}
	}

	override(&cfg)

	return &cfg, nil
}
