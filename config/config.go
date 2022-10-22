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
	JobName        string         `yaml:"job_name"`
	ScrapeInterval time.Duration  `yaml:"scrape_interval"`
	ScrapeTimeout  time.Duration  `yaml:"scrape_timeout"`
	MetricsPath    string         `yaml:"metrics_path"`
	Scheme         string         `yaml:"scheme"`
	StaticConfigs  []StaticConfig `yaml:"static_configs"`
}

type LogConfig struct {
	LogLevel string `yaml:"log_level"`
	LogFile  string `yaml:"log_file"`
}

type AutoIndexConfig struct {
	TiDBConfig      TiDBConfig      `yaml:"tidb"`
	NgMonitorConfig NgMonitorConfig `yaml:"ng-monitor"`
	WebConfig       WebConfig       `yaml:"web"`
	ScrapeConfigs   []*ScrapeConfig `yaml:"scrape_configs"`
	LogConfig       LogConfig       `yaml:"logs"`
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
