package common

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Algorithm AlgorithmConfig `yaml:"algorithm"`
	Replicas  []ReplicaConfig `yaml:"replicas"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type AlgorithmConfig struct {
	Type              string  `yaml:"type"`
	MaxLifeTime       int     `yaml:"maxLifeTime"`
	PoolSize          int     `yaml:"poolSize"`
	ProbeFactor       float64 `yaml:"probeFactor"`
	ProbeRemoveFactor int     `yaml:"probeRemoveFactor"`
	Mu                float64 `yaml:"mu"`
}

type ReplicaConfig struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Healthcheck string `yaml:"healthcheck"`
}

func LoadConfig(configpath string) Config {
	config, err := os.ReadFile(configpath)

	if err != nil {
		log.Panicf("could not load config. Please create a config.yaml file.")
	}

	var cfg Config
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		log.Panic(err)
	}

	return cfg
}
