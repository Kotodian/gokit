package utils

import (
	"os"

	"github.com/spf13/viper"
)

var Config *viper.Viper

// = NewConfigureFromEnv("CODEIN_APP_CONFIG", "yaml")

func InitConfigureFromEnv(env string, configType string) *viper.Viper {
	config := viper.New()
	config.SetConfigType(configType)
	config.SetConfigFile(os.Getenv(env))
	err := config.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return config
}
