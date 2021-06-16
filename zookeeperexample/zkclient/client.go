package main

import (
	"log"
	"time"
)

type ZkClient struct {
	servers        []string
	sessionTimeout time.Duration
}

func NewZkClient(servers []string, sessionTimeout time.Duration) *ZkClient {
	return &ZkClient{
		servers:        servers,
		sessionTimeout: sessionTimeout,
	}
}

func (zc *ZkClient) Start() error {
	zc.printf("Start ZkClient..")
	//conn, eventCh, err := zk.Connect(zc.servers, zc.sessionTimeout)
	return nil
}

func (zc *ZkClient) printf(format string, v ...interface{}) {
	log.Printf("[ZkClient]"+format, v)
}
