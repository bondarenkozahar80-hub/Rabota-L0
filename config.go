package config

import (
	"fmt"
	" os "
	" time "
	" os "
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	HTTP struct {
		Port string `yaml:"port"`
	} `yaml:"http"`
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
	Kafka struct {
		Brokers []string `yaml:"brokers"`
		Topic   string   `yaml:"topic"`
		GroupID string   `yaml:"group_id"`
	} `yaml:"kafka"`
}

func Load(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}
	return config, nil
}
