package main

import (
	"github.com/go-zookeeper/zk"
	"sync"
	"time"
)

type ZKClient struct {
	*zk.Conn
	servers        []string
	sessionTimeout time.Duration
}

func NewZKClient(servers []string, sessionTimeout time.Duration) (*ZKClient, error) {
	conn, eventChannel, err := zk.Connect(servers, sessionTimeout, zk.WithLogInfo(false))
	if err != nil {
		return nil, err
	}

	cli := ZKClient{
		Conn:           conn,
		servers:        servers,
		sessionTimeout: sessionTimeout,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-eventChannel:
				if e.Type != zk.EventSession {
					continue
				}

				switch e.State {
				case zk.StateConnected, zk.StateHasSession:
					return
				}
			}
		}
	}()
	wg.Wait()
	return &cli, nil
}
