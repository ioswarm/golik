package golik

import (
	"fmt"
	"github.com/spf13/viper"
	"strings"
	"time"
)

var GolikConfigFile string
var viperInit bool

func InitViperSettings() {
	if !viperInit {
		if GolikConfigFile != "" {
			viper.SetConfigFile(GolikConfigFile)
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



func newBaseSettings() *httpSettings {
	return &httpSettings{
		Host: viper.GetString("golik.http.host"),
		Port: viper.GetInt("golik.http.port"),
		ReadTimeout: time.Duration(viper.GetInt("golik.http.readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("golik.http.writeTimeout")) * time.Second,
		IdleTimeout: time.Duration(viper.GetInt("golik.http.idleTimeout")) * time.Second,
	}
}

func newSettings(name string) *httpSettings {
	bs := newBaseSettings()

	getPath := func(segment string) string {
		return fmt.Sprintf("golik.http.%v.%v", name, segment)
	} 

	path := getPath("host")
	if viper.IsSet(path) {
		bs.Host = viper.GetString(path)
	}
	
	path = getPath("port")
	if viper.IsSet(path) {
		bs.Port = viper.GetInt(path)
	}

	path = getPath("readTimeout")
	if viper.IsSet(path) {
		bs.ReadTimeout = time.Duration(viper.GetInt(path)) * time.Second
	}

	path = getPath("writeTimeout")
	if viper.IsSet(path) {
		bs.WriteTimeout = time.Duration(viper.GetInt(path)) * time.Second
	}

	path = getPath("idleTimeout")
	if viper.IsSet(path) {
		bs.IdleTimeout = time.Duration(viper.GetInt(path)) * time.Second
	}

	return bs
}

type httpSettings struct {
	Host string
	Port int
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	IdleTimeout time.Duration
}

func (s *httpSettings) Addr() string {
	return fmt.Sprintf("%v:%v", s.Host, s.Port)
}


func init() {
	
	// logging
	viper.SetDefault("golik.log.provider", "logrus")
	viper.SetDefault("golik.log.level", "INFO")
	viper.SetDefault("golik.log.formatter", "json")

	// http
	viper.SetDefault("golik.http.host", "")
	viper.SetDefault("golik.http.port", 9000)
	viper.SetDefault("golik.http.readTimeout", 5)
	viper.SetDefault("golik.http.writeTimeout", 10)
	viper.SetDefault("golik.http.idleTimeout", 15)
}