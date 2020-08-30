package http

import (
	"time"
	"fmt"

	"github.com/spf13/viper"
)



type httpSettings struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func (s *httpSettings) Addr() string {
	return fmt.Sprintf("%v:%v", s.Host, s.Port)
}


func newBaseHTTPSettings() *httpSettings {
	return &httpSettings{
		Host:         viper.GetString("http.host"),
		Port:         viper.GetInt("http.port"),
		ReadTimeout:  time.Duration(viper.GetInt("http.readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("http.writeTimeout")) * time.Second,
		IdleTimeout:  time.Duration(viper.GetInt("http.idleTimeout")) * time.Second,
	}
}

func newHTTPSettings(name string) *httpSettings {
	bs := newBaseHTTPSettings()

	getPath := func(segment string) string {
		return fmt.Sprintf("http.%v.%v", name, segment)
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


func init() {
	viper.SetDefault("http.host", "")
	viper.SetDefault("http.port", 9000)
	viper.SetDefault("http.readTimeout", 5)
	viper.SetDefault("http.writeTimeout", 10)
	viper.SetDefault("http.idleTimeout", 15)
}