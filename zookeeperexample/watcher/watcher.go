package main

import (
	"github.com/go-zookeeper/zk"
	"log"
	"strings"
)

type ZkWatcher struct {
	client *zk.Conn
	stop   chan bool
}

func (zw *ZkWatcher) watchDir(key string) {
	for {
		children, stat, eventCh, err := zw.client.ChildrenW(key)
		if err != nil {
			log.Printf("[ZkWatcher] failed to GetW. err: %v", err)
			return
		}
		log.Printf("[ZkWatcher] stats: %v, children: %s", stat, strings.Join(children, ","))

		select {
		case e := <-eventCh:
			log.Printf("[ZkWatcher] event occur: %v", e)
		case <-zw.stop:
			log.Printf("[ZkWatcher] terminate watchDir(%s)", key)
			return
		}
	}
}

func (zw *ZkWatcher) watchKey(key string) {
	for {
		contents, stat, eventCh, err := zw.client.GetW(key)
		if err != nil {
			log.Printf("[ZkWatcher] failed to GetW. err: %v", err)
			return
		}
		log.Printf("[ZkWatcher] stat: %v, contents: %s", stat, string(contents))

		select {
		case e := <-eventCh:
			log.Printf("[ZkWatcher] event occur: %v", e)
			if e.Type == zk.EventNodeDeleted {
				return
			}
		case <-zw.stop:
			log.Printf("[ZkWatcher] terminate watchKey(%s)", key)
			return
		}
	}
}

func (zw *ZkWatcher) Stop() {
	select {
	case <-zw.stop:
		return
	default:
		close(zw.stop)
	}
}
