package koanfexample

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

func Test2(t *testing.T) {
	val := envPrefix + "SERVERS_USERSERVER_WRITETIMEOUT"
	fmt.Println(strings.Replace(strings.ToLower(strings.TrimPrefix(val, envPrefix)), "_", ".", -1))
}

func TestLoad(t *testing.T) {
	configMap := map[string]interface{}{
		"servers.userserver.readTimeout":  "6s",
		"servers.userserver.writeTimeout": "4m",
		"servers.productserver.port":      8091,
	}

	configPath := "./test.yaml"

	os.Setenv(envPrefix+"SERVERS_PRODUCTSERVER_PORT", "8093")
	os.Setenv(envPrefix+"SERVERS_PRODUCTSERVER_READ-TIMEOUT", "2s")

	conf, err := Load(configMap, configPath)
	assert.NoError(t, err)

	data, err := json.MarshalIndent(&conf, "", "   ")
	if err != nil {
		log.Println("MarshalIndent ERR:", err)
		return
	}
	log.Printf(">> Configs: \n%s", data)
}

func TestValidate(t *testing.T) {
	configMap := map[string]interface{}{
		"servers.userserver.readTimeout": "notduration",
	}

	conf, err := Load(configMap, "")
	assert.Nil(t, conf)
	assert.Error(t, err)
	fmt.Println(err)
}
