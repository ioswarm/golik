package mongo

import (
	"fmt"
	"time"
	"github.com/spf13/viper"
)

type Settings struct {
	Host string
	Port int
	Collection string
	ConnectionTimeout time.Duration
	PingTimeout time.Duration
	CheckConnectionInterval time.Duration
}

func (s *Settings) URI() string {
	return fmt.Sprintf("mongodb://%v:%v", s.Host, s.Port)
}

func NewDefaultSettings() *Settings {
	return &Settings{
		Host: viper.GetString("golik.mongo.host"),
		Port: viper.GetInt("golik.mongo.port"),
		Collection: viper.GetString("golik.mongo.collection"),
		ConnectionTimeout: time.Duration(viper.GetInt("golik.mongo.connectionTimeout")) * time.Millisecond,
		PingTimeout: time.Duration(viper.GetInt("golik.mongo.pingTimeout")) * time.Millisecond,
		CheckConnectionInterval: time.Duration(viper.GetInt("golik.mongo.checkConnectionInterval")) * time.Millisecond,
	}
}

func NewSettings(name string) *Settings {
	ds := NewDefaultSettings()

	getPath := func(segment string) string {
		return fmt.Sprintf("golik.mongo.%v.%v", name, segment)
	} 

	path := getPath("host")
	if viper.IsSet(path) {
		ds.Host = viper.GetString(path)
	}

	path = getPath("port")
	if viper.IsSet(path) {
		ds.Port = viper.GetInt(path)
	}

	path = getPath("collection")
	if viper.IsSet(path) {
		ds.Collection = viper.GetString(path)
	}

	path = getPath("connectionTimeout")
	if viper.IsSet(path) {
		ds.ConnectionTimeout = time.Duration(viper.GetInt(path)) * time.Millisecond
	}

	path = getPath("pingTimeout") 
	if viper.IsSet(path) {
		ds.PingTimeout = time.Duration(viper.GetInt(path)) * time.Millisecond
	}

	path = getPath("checkConnectionInterval") 
	if viper.IsSet(path) {
		ds.CheckConnectionInterval = time.Duration(viper.GetInt(path)) * time.Millisecond
	}

	return ds
}

func init() {
	viper.SetDefault("golik.mongo.host", "localhost")
	viper.SetDefault("golik.mongo.port", "27017")
	viper.SetDefault("golik.mongo.collection", "golik")
	viper.SetDefault("golik.mongo.connectionTimeout", 5000)
	viper.SetDefault("golik.mongo.pingTimeout", 2000)
	viper.SetDefault("golik.mongo.checkConnectionInterval", 30000)
}