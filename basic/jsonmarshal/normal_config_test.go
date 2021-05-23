package jsonmarshal

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/jeremywohl/flatten"
	"log"
	"regexp"
	"strings"
	"testing"
)

type Config struct {
	Server ServerConfig `json:"server"`
	DB     DBConfig     `json:"db"`
	APM    APMConfig    `json:"apm"`
}

type ServerConfig struct {
	Port int `json:"port"`
}

type DBConfig struct {
	DSN         string `json:"dsn"`
	PoolMaxOpen int    `json:"poolMaxOpen"`
}

type APMConfig struct {
	License string `json:"license"`
}

func TestNormalConfig(t *testing.T) {
	c := &Config{
		Server: ServerConfig{
			Port: 8080,
		},
		DB: DBConfig{
			DSN:         "root:password@tcp(db)/local_db?charset=utf8&parseTime=True&multiStatements=true",
			PoolMaxOpen: 100,
		},
		APM: APMConfig{
			License: "my apm license!",
		},
	}
	value, err := toJsonWithMasking2(map[string]struct{}{
		"apm.license": {},
	}, c)
	if err != nil {
		log.Println("failed to marshal json", err)
		return
	}
	log.Printf(">> DisplayConfig\n%s", value)
	// Output
	//2021/05/13 22:47:30 >> DisplayConfig
	//{
	//    "apm.license": "****",
	//    "db.dsn": "root:****@tcp(db)/local_db?charset=utf8\u0026parseTime=True\u0026multiStatements=true",
	//    "db.poolMaxOpen": 100,
	//    "server.port": 8080
	//}
}

func toJsonWithMasking(maskKeys map[string]struct{}, cfg interface{}) (string, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	flat, err := flatten.FlattenString(string(data), "", flatten.DotStyle)
	if err != nil {
		return "", err
	}

	var m map[string]interface{}
	err = json.Unmarshal([]byte(flat), &m)
	if err != nil {
		return "", err
	}

	for key, val := range m {
		if v, ok := val.(string); ok {
			m[key] = maskPassword(v)
		}
		if _, ok := maskKeys[key]; ok {
			m[key] = "****"
		}
	}

	data, err = json.MarshalIndent(&m, "", "    ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func toJsonWithMasking2(maskKeys map[string]struct{}, cfg interface{}) (string, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	parsed, err := gabs.ParseJSON(data)
	if err != nil {
		return "", err
	}
	flat, err := parsed.Flatten()
	for key, val := range flat {
		if v, ok := val.(string); ok {
			flat[key] = maskPassword(v)
		}
		if _, ok := maskKeys[key]; ok {
			flat[key] = "****"
		}
	}

	data, err = json.MarshalIndent(&flat, "", "    ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func maskPassword(val string) string {
	regex := regexp.MustCompile(`^(?P<protocol>.+?\/\/)?(?P<username>.+?):(?P<password>.+?)@(?P<address>.+)$`)
	if !regex.MatchString(val) {
		return val
	}
	matches := regex.FindStringSubmatch(val)
	for i, v := range regex.SubexpNames() {
		if "password" == v {
			val = strings.ReplaceAll(val, matches[i], "***")
		}
	}
	return val
}

func TestMask(t *testing.T) {
	cases := []struct {
		Val string
	}{
		{
			Val: "root:password@tcp(db)/local_db?charset=utf8&parseTime=True&multiStatements=true",
		},
		{
			Val: "is not match",
		},
		{
			Val: "http://user1:pass@localhost:8080/path1/path2",
		},
	}

	for _, tc := range cases {
		fmt.Println("Origin:", tc.Val)
		fmt.Println("Replce:", maskPassword2(tc.Val))
		fmt.Println("--------------------------------")
	}
}

func maskPassword2(val string) string {
	return val
}
