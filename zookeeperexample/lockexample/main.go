package main

import (
	"fmt"
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
	c, _, err := zk.Connect(zkServers, time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	c.SetLogger(&zkLogger{})
	defer c.Close()
	cleanUpDataNodes(c)

	producerCount := 3
	mc := MessageCollector{}
	producers := make([]*Producer, producerCount)
	for i := 0; i < producerCount; i++ {
		producers[i] = NewProducer(c, &mc, fmt.Sprintf("Worker-%d", i+1))
		go producers[i].Start()
	}

	time.Sleep(time.Second * 10)
	for _, p := range producers {
		p.Stop()
	}
	time.Sleep(time.Second * 3)

	log.Println("Check message")
	log.Printf("> Received Messages: %d", len(mc.messages))
	messages := make(map[int]int)
	maxNumber := 0
	for _, m := range mc.messages {
		messages[m.Number] = messages[m.Number] + 1
		if maxNumber < m.Number {
			maxNumber = m.Number
		}
	}
	for k, v := range messages {
		log.Printf("> Message-%d > count: %d", k, v)
	}

	// check max message number
	log.Printf("Check workers number. current message number: %d", maxNumber)
	for _, p := range producers {
		log.Printf("> %s - %d", p.name, p.seq)
	}

	log.Println("Check children")
	children, _, err := c.Children("/message")
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range children {
		log.Printf("> Children - %s", c)
	}
}

func cleanUpDataNodes(client *zk.Conn) {
	log.Println("Cleanup /message children nodes")
	children, _, err := client.Children("/message")
	if err != nil {
		log.Println("failed to get children. err:", err)
		return
	}
	for _, c := range children {
		path := "/message/" + c
		_, stat, err := client.Get(path)
		if err != nil {
			log.Printf("failed to get %s", path)
		} else {
			err := client.Delete(path, stat.Version)
			log.Printf("Delete a node %s > %v", c, err)
		}
	}
}

type zkLogger struct {
}

func (l *zkLogger) Printf(foramt string, v ...interface{}) {
	//log.Printf("[Zookeeper]"+foramt, v)
}
