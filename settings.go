package golik

import (
	"strings"

	"github.com/spf13/viper"
)

var golikConfigFile string
var viperInit bool

func initSettings() {
	if !viperInit {
		if golikConfigFile != "" {
			viper.SetConfigFile(golikConfigFile)
		} else {
			viper.AddConfigPath(".")
			viper.AddConfigPath("conf")
			viper.AddConfigPath("configs")

			viper.SetConfigName("config")
		}

		viper.ReadInConfig()

		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()

		viperInit = true
	}
}