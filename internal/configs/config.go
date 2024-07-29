package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type resource struct {
	Name string
	Endpoint string
	Destination_URL string
}

type configuration struct {
	Server struct {
		Host string
		Listen_Port string
	}
	Resources []resource
}

var Config *configuration

func NewConfiguration() (*configuration, error) {
	viper.AddConfigPath("data")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error loading config file: %s", err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %s", err)
	}

	return Config, nil
}