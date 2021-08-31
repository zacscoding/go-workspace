package koanfexample

import (
	"encoding/json"
	"fmt"
	"github.com/huandu/xstrings"
	"github.com/jeremywohl/flatten"
	"github.com/knadh/koanf"
	kjson "github.com/knadh/koanf/parsers/json"
	kyaml "github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/pkg/errors"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const envPrefix = "APP_"

type Conf struct {
	*koanf.Koanf
	Servers map[string]ServerConf `json:"servers" validate:"dive"`
	DBConf  DBConf                `json:"db"`
}

func (c *Conf) Validate() error {
	for k, s := range c.Servers {
		if err := s.Validate(); err != nil {
			return errors.Wrap(err, "servers."+k)
		}
	}

	if err := c.DBConf.Validate(); err != nil {
		return errors.Wrap(err, "db")
	}
	return nil
}

type ServerConf struct {
	Port            int           `json:"port"`
	ReadTimeoutVal  string        `json:"readTimeout"`
	ReadTimeout     time.Duration `json:"-"`
	WriteTimeoutVal string        `json:"writeTimeout"`
	WriteTimeout    time.Duration `json:"-"`
}

func (c *ServerConf) Validate() error {
	var err error
	c.ReadTimeout, err = time.ParseDuration(c.ReadTimeoutVal)
	if err != nil {
		return errors.Wrap(err, "readTimeout")
	}
	c.WriteTimeout, err = time.ParseDuration(c.WriteTimeoutVal)
	if err != nil {
		return errors.Wrap(err, "writeTimeout")
	}
	return nil
}

type DBConf struct {
	DataSourceName string `json:"dataSourceName"`
	Pool           struct {
		MaxOpen        int           `json:"maxOpen"`
		MaxIdle        int           `json:"maxIdle"`
		MaxLifetimeVal string        `json:"maxLifetime"`
		MaxLifetime    time.Duration `json:"-"`
	} `json:"pool"`
}

func (c *DBConf) Validate() error {
	var err error
	c.Pool.MaxLifetime, err = time.ParseDuration(c.Pool.MaxLifetimeVal)
	if err != nil {
		return errors.Wrap(err, "pool.maxLifetime")
	}
	return nil
}

// MarshalJSON returns a flat json data with masking values such as db password or jwt.secret config.
func (c *Conf) MarshalJSON() ([]byte, error) {
	conf := struct {
		Servers map[string]ServerConf `json:"servers"`
		DBConf  DBConf                `json:"db"`
	}{
		Servers: c.Servers,
		DBConf:  c.DBConf,
	}
	data, err := json.Marshal(&conf)
	if err != nil {
		return nil, err
	}
	flat, err := flatten.FlattenString(string(data), "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = json.Unmarshal([]byte(flat), &m)
	if err != nil {
		return nil, err
	}

	maskKeys := map[string]struct{}{
		// add keys if u want to mask some properties.
	}

	for key, val := range m {
		if v, ok := val.(string); ok {
			m[key] = maskPassword(v)
		}
		if _, ok := maskKeys[key]; ok {
			m[key] = "****"
		}
	}
	return json.Marshal(&m)
}

// Load returns a new Conf from given configMap and configPath with below order.
// (1) load defaults.
// (2) load given configMap.
// (3) load configPath.
// (4) load env.
func Load(configMap map[string]interface{}, configPath string) (*Conf, error) {
	k := koanf.New(".")

	// ------------------------------------------------------
	// (1) Load defaults.
	// ------------------------------------------------------
	if err := k.Load(confmap.Provider(defaultConf, "."), nil); err != nil {
		return nil, errors.Wrap(err, "load from defaults")
	}

	// ------------------------------------------------------
	// (2) load config map if exists.
	// ------------------------------------------------------
	if len(configMap) != 0 {
		if err := k.Load(confmap.Provider(configMap, "."), nil); err != nil {
			return nil, errors.Wrap(err, "load from config map")
		}
	}

	// ------------------------------------------------------
	// (3) load configPath if exists.
	// ------------------------------------------------------
	if configPath != "" {
		path, err := filepath.Abs(configPath)
		if err != nil {
			return nil, errors.Wrap(err, "invalid configPath")
		}
		var parser koanf.Parser
		switch ext := filepath.Ext(configPath); ext {
		case ".yaml", ".yml":
			parser = kyaml.Parser()
		case ".json":
			parser = kjson.Parser()
		default:
			return nil, fmt.Errorf("not supported file extension: %s", ext)
		}
		if err := k.Load(file.Provider(path), parser); err != nil {
			return nil, errors.Wrap(err, "load from config path")
		}
	}

	// ------------------------------------------------------
	// (4) load from env.
	// ------------------------------------------------------
	if err := k.Load(env.Provider(envPrefix, ".", func(s string) string {
		// Remove env prefix and convert to lowercase, replace "_" to ".".
		// e.g: APP_SERVERS_USERSERVER_READ-TIMEOUT => servers.userserver.read-timeout
		replaced := strings.Replace(strings.ToLower(strings.TrimPrefix(s, envPrefix)), "_", ".", -1)
		// Convert kebab to camel with first rune to lower.
		// e.g: servers.userserver.read-timeout => servers.userserver.readTimeout
		return xstrings.FirstRuneToLower(xstrings.ToCamelCase(replaced))
	}), nil); err != nil {
		return nil, errors.Wrap(err, "load from env")
	}

	var conf Conf
	if err := k.UnmarshalWithConf("", &conf, koanf.UnmarshalConf{Tag: "json", FlatPaths: false}); err != nil {
		return nil, errors.Wrap(err, "unmarshal with conf")
	}
	conf.Koanf = k
	if err := conf.Validate(); err != nil {
		return nil, errors.Wrap(err, "validate from conf")
	}
	return &conf, nil
}

func maskPassword(val string) string {
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
