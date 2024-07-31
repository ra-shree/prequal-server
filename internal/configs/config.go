package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type resource struct {
	Name            string
	Endpoint        string
	Destination_Url string
}

type Configuration struct {
	Server struct {
		Host        string
		Listen_Port string
	}
	Resources []resource
}

var Config *Configuration

func NewConfiguration() (*Configuration, error) {
	viper.AddConfigPath("data")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error loading config file: %s", err)
	}

	if err := viper.Unmarshal(&Config); err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	return Config, nil
}
