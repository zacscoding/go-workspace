package main

import (
	"github.com/go-zookeeper/zk"
	"log"
	"time"
)

const (
	enableZkLog    = false
	enableEventLog = false
)

var (
	zkServers = []string{
		"localhost:2181",
	}
)

func main() {
	c, eventCh, err := zk.Connect(zkServers, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	c.SetLogger(&zkLogger{})
	defer c.Close()
	go func() {
		for {
			select {
			case e := <-eventCh:
				if enableEventLog {
					log.Println("[EventLoop] EventOccur", e)
				}
			}
		}
	}()

	path := "/MyFirstZnode"
	// Check "/MyFirstZnode" and Delete if exist
	ok, stat, err := c.Exists(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("c.Exists > ok: %v, stat: %v", ok, stat)
	if ok {
		if err := c.Delete(path, stat.Version); err != nil {
			log.Fatal(err)
		}
	}
	result, err := c.Create(path, []byte("MyData"), 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("c.Create > %s", result)

	result, err = c.Create(path, []byte("MyData"), 0, zk.WorldACL(zk.PermAll))
	if err != zk.ErrNodeExists {
		log.Fatal(err)
	}
}

type zkLogger struct {
}

func (l *zkLogger) Printf(foramt string, v ...interface{}) {
	if enableZkLog {
		log.Printf("[Zookeeper]"+foramt, v)
	}
}
