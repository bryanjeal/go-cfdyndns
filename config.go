package main

import (
	"flag"
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/spf13/viper"
)

type appConfiguration struct {
	mutex *sync.RWMutex
}

func (c *appConfiguration) GetBool(key string) bool {
	c.mutex.RLock()
	value := viper.GetBool(key)
	c.mutex.RUnlock()

	return value
}
func (c *appConfiguration) GetInt(key string) int {
	c.mutex.RLock()
	value := viper.GetInt(key)
	c.mutex.RUnlock()

	return value
}
func (c *appConfiguration) GetString(key string) string {
	c.mutex.RLock()
	value := viper.GetString(key)
	c.mutex.RUnlock()

	return value
}
func (c *appConfiguration) GetStringSlice(key string) []string {
	c.mutex.RLock()
	value := viper.GetStringSlice(key)
	c.mutex.RUnlock()

	return value
}

func (c *appConfiguration) Set(key string, value interface{}) {
	c.mutex.Lock()
	viper.Set(key, value)
	c.mutex.Unlock()
}
func configureApp(configName *string) *appConfiguration {

	appConf := &appConfiguration{
		mutex: &sync.RWMutex{},
	}

	// configure app
	flag.Parse()
	viper.AutomaticEnv()

	// defaults: logging
	viper.SetDefault("log.mkdir", true)
	viper.SetDefault("log.toStderr", false)
	viper.SetDefault("log.alsoToStderr", false)
	viper.SetDefault("log.verbosity", 0)
	viper.SetDefault("log.stderrThreshold", "ERROR")
	viper.SetDefault("log.dir", "./log/")

	// find config file
	if configName != nil && len(*configName) > 0 {
		viper.SetConfigName(*configName)
	} else {
		viper.SetConfigName("config")
	}
	viper.AddConfigPath(".")
	viper.AddConfigPath("/usr/local/etc/" + app + "/")
	viper.AddConfigPath("/etc/" + app + "/")
	viper.AddConfigPath("$HOME/." + app)

	// read config file
	err := viper.ReadInConfig()
	if err != nil {
		viper.SetConfigName(app)
		viper.AddConfigPath("/usr/local/etc/")
		viper.AddConfigPath("/etc/")
		err = viper.ReadInConfig()
		if err != nil {
			flag.Set("alsologtostderr", "true")
			glog.Errorf("error config file: %s \nUsing Default Config\n", err)
		}
	}

	// mkdir for logging if we need to
	if viper.GetBool("log.mkdir") {
		os.MkdirAll(viper.GetString("log.dir"), 0755)
	}

	// set glog flags from config
	flag.Set("logtostderr", viper.GetString("log.toStderr"))
	flag.Set("alsologtostderr", viper.GetString("log.alsoToStderr"))
	flag.Set("v", viper.GetString("log.verbosity"))
	flag.Set("stderrthreshold", viper.GetString("log.stderrThreshold"))
	flag.Set("log_dir", viper.GetString("log.dir"))

	return appConf
}
