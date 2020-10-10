package golik

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

var golikConfigFile string
var viperInit bool

type CloveSettings interface {
	BufferSize() uint32

	PreStartTimeout() time.Duration
	PostStartTimeout() time.Duration
	PreStopTimeout() time.Duration
	PostStopTimeout() time.Duration

	SubscriptionTimeout() time.Duration
}

func newCloveSettings(name string) CloveSettings {
	prefix := "golik.clove."

	cs := &cloveSettings{
		bufferSize:          viper.GetUint32(prefix + "bufferSize"),
		preStartTimeout:     viper.GetDuration(prefix+"preStartTimeout") * time.Second,
		postStartTimeout:    viper.GetDuration(prefix+"postStartTimeout") * time.Second,
		preStopTimeout:      viper.GetDuration(prefix+"preStopTimeout") * time.Second,
		postStopTimeout:     viper.GetDuration(prefix+"postStopTimeout") * time.Second,
		subscriptionTimeout: viper.GetDuration(prefix+"subscriptionTimeout") * time.Second,
	}

	if name != "" {
		prefix = prefix + name + "."

		scs := &cloveSettings{
			bufferSize:          cs.bufferSize,
			preStartTimeout:     cs.preStartTimeout,
			postStartTimeout:    cs.postStartTimeout,
			preStopTimeout:      cs.preStopTimeout,
			postStopTimeout:     cs.postStopTimeout,
			subscriptionTimeout: cs.subscriptionTimeout,
		}

		if viper.IsSet(prefix + "buffersize") {
			scs.bufferSize = viper.GetUint32(prefix + "bufferSize")
		}
		if viper.IsSet(prefix + "preStartTimeout") {
			scs.preStartTimeout = viper.GetDuration(prefix+"preStartTimeout") * time.Second
		}
		if viper.IsSet(prefix + "postStartTimeout") {
			scs.postStartTimeout = viper.GetDuration(prefix+"postStartTimeout") * time.Second
		}
		if viper.IsSet(prefix + "preStopTimeout") {
			scs.preStopTimeout = viper.GetDuration(prefix+"preStopTimeout") * time.Second
		}
		if viper.IsSet(prefix + "postStopTimeout") {
			scs.postStopTimeout = viper.GetDuration(prefix+"postStopTimeout") * time.Second
		}
		if viper.IsSet(prefix + "subscriptionTimeout") {
			scs.subscriptionTimeout = viper.GetDuration(prefix+"subscriptionTimeout") * time.Second
		}

		return scs
	}

	return cs
}

type cloveSettings struct {
	bufferSize uint32

	preStartTimeout  time.Duration
	postStartTimeout time.Duration
	preStopTimeout   time.Duration
	postStopTimeout  time.Duration

	subscriptionTimeout time.Duration
}

func (c *cloveSettings) BufferSize() uint32 {
	return c.bufferSize
}
func (c *cloveSettings) PreStartTimeout() time.Duration {
	return c.preStartTimeout
}
func (c *cloveSettings) PostStartTimeout() time.Duration {
	return c.postStartTimeout
}
func (c *cloveSettings) PreStopTimeout() time.Duration {
	return c.preStopTimeout
}
func (c *cloveSettings) PostStopTimeout() time.Duration {
	return c.postStopTimeout
}
func (c *cloveSettings) SubscriptionTimeout() time.Duration {
	return c.subscriptionTimeout
}

type Settings interface {
	TerminationTimeout() time.Duration

	DefaultCloveSettings() CloveSettings

	CloveSettings(name string) CloveSettings
}

func NewSettings() Settings {
	initSettings()
	return &settings{
		terminationTimeout:   viper.GetDuration("golik.terminationTimeout") * time.Second,
		defaultCloveSettings: newCloveSettings(""),
	}
}

type settings struct {
	terminationTimeout   time.Duration
	defaultCloveSettings CloveSettings
}

func (s *settings) TerminationTimeout() time.Duration {
	return s.terminationTimeout
}

func (s *settings) DefaultCloveSettings() CloveSettings {
	return s.defaultCloveSettings
}

func (s *settings) CloveSettings(name string) CloveSettings {
	return newCloveSettings(name)
}

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

func init() {
	viper.SetDefault("golik.terminationTimeout", 30)

	viper.SetDefault("golik.clove.bufferSize", 1000)
	viper.SetDefault("golik.clove.preStartTimeout", 10)
	viper.SetDefault("golik.clove.postStartTimeout", 10)
	viper.SetDefault("golik.clove.preStopTimeout", 10)
	viper.SetDefault("golik.clove.postStopTimeout", 10)
	viper.SetDefault("golik.clove.subscriptionTimeout", 10)
}
