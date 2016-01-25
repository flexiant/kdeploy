package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Connection struct {
		APIEndpoint string
		CACert      string
		Cert        string
		Key         string
	}
}

func ReadConfig() (*Config, error) {
	cfg, err := ReadConfigFrom(".kdeploy.yml")
	if err == nil {
		return cfg, nil
	}
	cfg, err = ReadConfigFrom("~/.kdeploy.yml")
	if err == nil {
		return cfg, nil
	}
	return nil, fmt.Errorf("could not read config")
}

func ReadConfigFrom(path string) (*Config, error) {
	cfgFile, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	configBytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(configBytes, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
