package jsonmarshal

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
)

type FlattenConfig struct {
	ServerPort    int    `json:"server.port"`
	DBDSN         string `json:"db.dsn"`
	DBPoolMaxOpen int    `json:"db.poolMaxOpen"`
	APMLicense    string `json:"apm.license"`
}

func TestFlattenConfig(t *testing.T) {
	c := &FlattenConfig{
		ServerPort:    8080,
		DBDSN:         "root:password@tcp(db)/local_db?charset=utf8&parseTime=True&multiStatements=true",
		DBPoolMaxOpen: 100,
		APMLicense:    "my apm license!",
	}
	value, err := toFlattenJsonWithMasking(map[string]struct{}{
		"apm.license": {},
	}, c)
	if err != nil {
		log.Println("failed to marshal json", err)
		return
	}
	log.Printf(">> DisplayConfig\n%s", value)
	// Output
	//{
	//    "apm.license": "****",
	//    "db.dsn": "root:****@tcp(db)/local_db?charset=utf8\u0026parseTime=True\u0026multiStatements=true",
	//    "db.poolMaxOpen": 100,
	//    "server.port": 8080
	//}
}

func toFlattenJsonWithMasking(maskKeys map[string]struct{}, cfg interface{}) (string, error) {
	elts := reflect.ValueOf(cfg).Elem()
	fieldsLen := elts.NumField()
	m := make(map[string]interface{}, fieldsLen)

	for i := 0; i < fieldsLen; i++ {
		f := elts.Field(i)
		jsonKey := elts.Type().Field(i).Tag.Get("json")
		if jsonKey == "" {
			continue
		}

		m[jsonKey] = f.Interface()
		if v, ok := m[jsonKey].(string); ok {
			m[jsonKey] = maskPassword(v)
		}
		if _, ok := maskKeys[jsonKey]; ok {
			m[jsonKey] = "****"
		}
	}
	data, err := json.MarshalIndent(&m, "", "    ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
