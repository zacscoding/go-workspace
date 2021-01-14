package timeoutexample

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestReverseProxy(t *testing.T) {
	var (
		// Proxy Server
		serverReadTimeout  = 5 * time.Second
		serverWriteTimeout = 10 * time.Second
		proxyDialerTimeout = 20 * time.Second
		// http client
		cliTimeout = 20 * time.Second
		sleepSec   = "15"
	)

	// 1) start remote server
	go startRemoteServer(":8000")
	go startProxyServer(":8890", "http://localhost:8000", serverReadTimeout, serverWriteTimeout, proxyDialerTimeout)

	// 2) define call func
	httpCallFunc := func(path string, cliTimeout time.Duration, sleep string, body map[string]interface{}) error {
		url := fmt.Sprintf("http://localhost:8890%s", path)
		if sleep != "" {
			url += "?sleep=" + sleep
		}

		bodyBytes, _ := json.Marshal(&body)
		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Request-Id", uuid.New().String())

		cli := &http.Client{
			Timeout: cliTimeout,
		}
		resp, err := cli.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		respBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[Client] success to call. status code: %d, response body:%s", resp.StatusCode, string(respBytes))
		return nil
	}

	// 1) call
	start := time.Now()
	err := httpCallFunc("/reverse", cliTimeout, sleepSec, map[string]interface{}{
		"call": "call1",
	})
	log.Printf("[Client] request [%s], err:%v\n", time.Since(start), err)

	// 2) call
	start = time.Now()
	err = httpCallFunc("/not-exist", cliTimeout, sleepSec, map[string]interface{}{
		"call": "call1",
	})
	log.Printf("[Client] request [%s], err:%v\n", time.Since(start), err)
}
