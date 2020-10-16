package golik

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type httpSettings struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

func (s *httpSettings) Addr() string {
	return fmt.Sprintf("%v:%v", s.Host, s.Port)
}

func newBaseHTTPSettings() *httpSettings {
	return &httpSettings{
		Host:            viper.GetString("http.host"),
		Port:            viper.GetInt("http.port"),
		ReadTimeout:     viper.GetDuration("http.readTimeout") * time.Second,
		WriteTimeout:    viper.GetDuration("http.writeTimeout") * time.Second,
		IdleTimeout:     viper.GetDuration("http.idleTimeout") * time.Second,
		ShutdownTimeout: viper.GetDuration("http.shutdownTimeout") * time.Second,
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
		bs.ReadTimeout = viper.GetDuration(path) * time.Second
	}

	path = getPath("writeTimeout")
	if viper.IsSet(path) {
		bs.WriteTimeout = viper.GetDuration(path) * time.Second
	}

	path = getPath("idleTimeout")
	if viper.IsSet(path) {
		bs.IdleTimeout = viper.GetDuration(path) * time.Second
	}

	path = getPath("shutdownTimeout")
	if viper.IsSet(path) {
		bs.ShutdownTimeout = viper.GetDuration(path) * time.Second
	}

	return bs
}

func init() {
	viper.SetDefault("http.host", "")
	viper.SetDefault("http.port", 9000)
	viper.SetDefault("http.readTimeout", 5)
	viper.SetDefault("http.writeTimeout", 10)
	viper.SetDefault("http.idleTimeout", 15)
	viper.SetDefault("http.shutownTimeout", 10)
}
