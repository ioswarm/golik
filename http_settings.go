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
		Host:            viper.GetString("golik.http.host"),
		Port:            viper.GetInt("golik.http.port"),
		ReadTimeout:     viper.GetDuration("golik.http.readTimeout") * time.Second,
		WriteTimeout:    viper.GetDuration("golik.http.writeTimeout") * time.Second,
		IdleTimeout:     viper.GetDuration("golik.http.idleTimeout") * time.Second,
		ShutdownTimeout: viper.GetDuration("golik.http.shutdownTimeout") * time.Second,
	}
}

func newHTTPSettings(name string) *httpSettings {
	bs := newBaseHTTPSettings()

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
	viper.SetDefault("golik.http.host", "")
	viper.SetDefault("golik.http.port", 9000)
	viper.SetDefault("golik.http.readTimeout", 5)
	viper.SetDefault("golik.http.writeTimeout", 10)
	viper.SetDefault("golik.http.idleTimeout", 15)
	viper.SetDefault("golik.http.shutownTimeout", 10)
}
