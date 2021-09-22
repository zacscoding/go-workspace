package main

import (
	"encoding/json"
	"github.com/knadh/koanf"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	C       *koanf.Koanf `json:"-" yaml:"-"`
	Stage   string       `json:"stage" yaml:"stage"`
	Logging LogConfig    `json:"logging" yaml:"logging"`
	Server  ServerConfig `json:"server" yaml:"server"`
	DB      DBConfig     `json:"db" yaml:"db"`
}

type LogConfig struct {
	Level    int    `json:"level" yaml:"level"`
	Encoding string `json:"encoding" yaml:"encoding"`
}

type ServerConfig struct {
	Port         int           `json:"port" yaml:"port"`
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout"`
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout"`
}

type DBConfig struct {
	DataSourceName string `json:"data-source-name" yaml:"data-source-name"`
	Migrate        struct {
		Enable bool   `json:"enable" yaml:"enable"`
		Dir    string `json:"dir" yaml:"dir"`
	} `json:"migrate" yaml:"migrate"`
	Pool struct {
		MaxOpen     int           `json:"max-open" yaml:"max-open"`
		MaxIdle     int           `json:"max-idle" yaml:"max-idle"`
		MaxLifetime time.Duration `json:"max-lifetime" yaml:"max-lifetime"`
	} `json:"pool" yaml:"pool"`
}

func (c *Config) MarshalJSON() ([]byte, error) {
	var (
		maskingKeys = map[string]struct{}{
			"server.port": {},
		}
		keys = c.C.Keys()
		m    = make(map[string]interface{}, len(keys))
	)

	for _, key := range keys {
		value := c.C.Get(key)
		if v, ok := value.(string); ok {
			value = maskPassword(v)
		}
		if _, ok := maskingKeys[key]; ok {
			value = "****"
		}
		m[key] = value
	}
	return json.Marshal(&m)
}

func maskPassword(val string) string {
	// TODO: change expression(if include @ in password)
	regex := regexp.MustCompile(`^(?P<protocol>.+?//)?(?P<username>.+?):(?P<password>.+?)@(?P<address>.+)$`)
	if !regex.MatchString(val) {
		return val
	}
	matches := regex.FindStringSubmatch(val)
	for i, v := range regex.SubexpNames() {
		if "password" == v {
			val = strings.ReplaceAll(val, matches[i], "****")
		}
	}
	return val
}
