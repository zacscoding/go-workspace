package main

import (
	"encoding/json"
	"fmt"
	"github.com/knadh/koanf"
	kjson "github.com/knadh/koanf/parsers/json"
	kyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	appPrefix = "MYAPP_"
)

var configFile string

func main() {
	configFile = "./poc/config/config.yaml"
	// configFile = "./poc/config/config.json"

	os.Setenv("MYAPP_STAGE", "env")
	os.Setenv("MYAPP_SERVER_READ-TIMEOUT", "1s")
	os.Setenv("MYAPP_SERVER_ENDPOINTS", "env-1,env-2")

	conf, err := load()
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.MarshalIndent(&conf, "", "    ")
	log.Printf("Load configs\n%s", string(b))
}

func load() (*Config, error) {
	k := koanf.New(".")
	// 1) load defaults
	if err := loadDefaults(k); err != nil {
		return nil, errors.Wrap(err, "load defaults")
	}
	// 2) load from file
	if err := loadFromFile(k, configFile); err != nil {
		return nil, errors.Wrap(err, "load from file")
	}
	// 3) load from env
	if err := loadFromEnv(k); err != nil {
		return nil, errors.Wrap(err, "load from env")
	}

	conf := Config{C: k}
	if err := k.UnmarshalWithConf("", &conf, koanf.UnmarshalConf{Tag: "json", FlatPaths: false}); err != nil {
		return nil, errors.Wrap(err, "unmarshal conf")
	}
	conf.C = k
	return &conf, nil
}

func loadDefaults(k *koanf.Koanf) error {
	return k.Load(confmap.Provider(map[string]interface{}{
		"stage": "local",

		"logging.level":    0,
		"logging.encoding": "json",

		"server.port":          8080,
		"server.read-timeout":  "10s",
		"server.write-timeout": "5m",
		"server.endpoints":     []string{"default-1", "default-2"},

		"db.data-source-name":  "root:password@tcp(db)/local_db?charset=utf8&parseTime=True&multiStatements=true",
		"db.migrate.enabled":   false,
		"db.migrate.dir":       "migrations/",
		"db.pool.max-open":     50,
		"db.pool.max-idle":     50,
		"db.pool.max-lifetime": "86400s",
	}, "."), nil)
}

func loadFromFile(k *koanf.Koanf, configFile string) error {
	path, err := filepath.Abs(configFile)
	if err != nil {
		return errors.Wrap(err, "convert to absolute path")
	}
	var (
		parser koanf.Parser
		ext    = filepath.Ext(path)
	)
	switch ext {
	case ".yaml", ".yml":
		parser = kyaml.Parser()
	case ".json":
		parser = kjson.Parser()
	default:
		return fmt.Errorf("not supported config file extension: %s. full path: %s", ext, configFile)
	}
	return k.Load(file.Provider(path), parser)
}

func loadFromEnv(k *koanf.Koanf) error {
	return k.Load(env.ProviderWithValue(appPrefix, ".", func(key string, value string) (string, interface{}) {
		key = strings.Replace(strings.ToLower(strings.TrimPrefix(key, appPrefix)), "_", ".", -1)
		switch k.Get(key).(type) {
		case []interface{}, []string:
			return key, strings.Split(value, ",")
		}
		return key, value
	}), nil)
}
