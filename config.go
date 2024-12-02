package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	_ "log"
)

type Config struct {
	Grafana struct {
		APIToken string `yaml:"api_token"`
		URL      string `yaml:"url"`
	} `yaml:"grafana"`
	Slack struct {
		APIToken string `yaml:"api_token"`
		Channel  string `yaml:"channel"`
	} `yaml:"slack"`
	Snapshot struct {
		Dashboards []string `yaml:"dashboards"`
		Expires    int      `yaml:"expires"`
	} `yaml:"snapshot"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
