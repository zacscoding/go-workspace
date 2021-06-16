package main

import (
	"github.com/go-zookeeper/zk"
	"log"
	"time"
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
	defer c.Close()
	go func() {
		for {
			select {
			case e := <-eventCh:
				log.Println("EventOccur", e)
			}
		}
	}()
	zk.NewLock(c, "path", zk.WorldACL(zk.PermAll))

}
