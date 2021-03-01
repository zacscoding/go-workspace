package timeoutexample

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestReverseProxy_MaxConns(t *testing.T) {
	var (
		// Proxy Server
		serverReadTimeout  = 5 * time.Second
		serverWriteTimeout = 15 * time.Second
		proxyDialerTimeout = 10 * time.Second
		// http client
		cliTimeout = 30 * time.Second
		sleepSec   = "20"
	)

	// 1) start remote server
	go startRemoteServer(false, ":8000")
	go startProxyServer(false, ":8890", "http://localhost:8000", serverReadTimeout, serverWriteTimeout, proxyDialerTimeout)

	requestFunc := func(name string) {
		url := "http://localhost:8890/reverse"
		if sleepSec != "" {
			url += "?sleep=" + sleepSec
		}
		body := map[string]interface{}{
			"call": "call1",
		}
		bodyBytes, _ := json.Marshal(&body)
		req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			log.Printf("[Client-%s] failed to request. err:%v\n", name, err)
			return
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("X-Request-Id", uuid.New().String())

		cli := &http.Client{
			Timeout: cliTimeout,
		}
		elapsed := time.Now()
		resp, err := cli.Do(req)
		if err != nil {
			log.Printf("[Client-%s] failed to request[%s]. err:%v\n", name, time.Now().Sub(elapsed), err)
			return
		}
		defer resp.Body.Close()
		respBytes, _ := ioutil.ReadAll(resp.Body)
		log.Printf("[Client-%s] success to call[%s]. status code: %d, response body:%s\n", name, time.Now().Sub(elapsed), resp.StatusCode, string(respBytes))
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			requestFunc(fmt.Sprintf("request-%d", i))
		}()
	}
	wg.Wait()
}

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
	// go startRemoteServer(true, ":8000")
	go startProxyServer(true, ":8890", "http://localhost:8000", serverReadTimeout, serverWriteTimeout, proxyDialerTimeout)

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
