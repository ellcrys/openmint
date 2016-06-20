package config

import (
	"strconv"
)

var (
	C = &Config{ make(map[string]interface{}) }
)

type Config struct {
	c map[string]interface{}
}

func(conf *Config) Add(k string, v interface{}) {
	conf.c[k] = v
}

func(conf *Config) Get(k string) interface{} {
	return conf.c[k]
}

func(conf *Config) GetString(k string) string {
	return conf.c[k].(string)
}

func (conf *Config) GetInt(k string) int {
	i, err := strconv.ParseInt(conf.c[k].(string), 10, 64)
	if err != nil {
		panic(err)
	}
	return int(i)
}

func (conf *Config) All() map[string]interface{} {
	return conf.c
}
