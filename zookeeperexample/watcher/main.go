package main

import (
	"github.com/go-zookeeper/zk"
	"log"
	"time"
)

const (
	enableZkLog        = false
	enableEventLoopLog = false
)

var (
	zkServers = []string{
		"localhost:2181",
	}
	lockPath = "/mylock"
)

func main() {
	client, eventCh, err := zk.Connect(zkServers, time.Second)
	if err != nil {
		log.Fatal(err)
	}
	client.SetLogger(&zkLogger{})
	defer client.Close()
	go func() {
		for {
			select {
			case e := <-eventCh:
				if enableEventLoopLog {
					log.Println("[EventLoop] EventOccur", e)
				}
			}
		}
	}()
	defer client.Close()
	watcher := &ZkWatcher{
		client: client,
		stop:   make(chan bool),
	}
	go watcher.watchDir(lockPath)

	l := zk.NewLock(client, lockPath, zk.WorldACL(zk.PermAll))
	err = l.Lock()
	log.Println("> Acquire lock. err:", err)
	time.Sleep(time.Second * 5)
	err = l.Unlock()
	log.Println("> Release lock. err:", err)
	time.Sleep(time.Second * 5)
	watcher.Stop()
}

type zkLogger struct {
}

func (l *zkLogger) Printf(foramt string, v ...interface{}) {
	if enableZkLog {
		log.Printf("[Zookeeper]"+foramt, v)
	}
}
