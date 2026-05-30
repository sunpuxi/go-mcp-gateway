package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Admin    AdminConfig    `yaml:"admin"`
	Database DatabaseConfig `yaml:"database"`
}

type AdminConfig struct {
	APIKey string `yaml:"api_key"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 环境变量覆盖（Docker / 生产环境部署用）
	if v := os.Getenv("SERVER_HOST"); v != "" {
		cfg.Server.Host = v
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = p
		}
	}
	if v := os.Getenv("DB_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("ADMIN_API_KEY"); v != "" {
		cfg.Admin.APIKey = v
	}

	return &cfg, nil
}
