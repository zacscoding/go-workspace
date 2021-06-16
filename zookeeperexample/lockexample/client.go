package main

import (
	"fmt"
	"github.com/go-zookeeper/zk"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Message struct {
	Number int
}

type Producer struct {
	client *zk.Conn
	mc     *MessageCollector
	name   string
	seq    int
	stop   chan bool
}

func NewProducer(client *zk.Conn, mc *MessageCollector, name string) *Producer {
	if ok, _, _ := client.Exists("/message"); !ok {
		client.Create("/message", []byte{}, 0, zk.WorldACL(zk.PermAll))
	}
	return &Producer{
		client: client,
		mc:     mc,
		name:   name,
		seq:    0,
		stop:   make(chan bool),
	}
}

func (p *Producer) Start() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sleepMills(rand.Intn(1000))
			p.seq++
			if err := p.produceMessage(); err != nil {
				log.Printf("[Producer-%s] failed to produce message: %d", p.name, p.seq)
			}
		case <-p.stop:
			return
		}
	}
}

func (p *Producer) Stop() {
	select {
	case <-p.stop:
		return
	default:
		close(p.stop)
	}
}

func (p *Producer) produceMessage() error {
	lock := NewLock(p.client, fmt.Sprintf("/mylock/%d", p.seq), zk.WorldACL(zk.PermAll))
	var (
		messagePath = fmt.Sprintf("/message/%d", p.seq)
		lastErr     error
	)
	defer lock.Unlock()
	for i := 0; i < 5; i++ {
		// check already proceed message.
		ok, _, _ := p.client.Exists(messagePath)
		if ok {
			return nil
		}
		// try to acquire a lock
		err := lock.Lock(time.Second)
		if err == nil {
			// check again already proceed message.
			if ok, _, _ := p.client.Exists(messagePath); ok {
				return nil
			}
			sleepMills(rand.Intn(500))
			// publish message
			p.mc.collect(Message{Number: p.seq})
			// create "/message/{number" node.
			_, err = p.client.Create(messagePath, []byte(p.name), 0, zk.WorldACL(zk.PermAll))
			if err == nil {
				return nil
			}
		}
		ok, _, err = p.client.Exists(messagePath)
		if ok {
			return nil
		}
		lastErr = err
		time.Sleep(time.Millisecond * 200)
	}
	return lastErr
}

func sleepMills(duration int) {
	time.Sleep(time.Duration(duration) * time.Millisecond)
}

type MessageCollector struct {
	messages []Message
	mutex    sync.Mutex
}

func (mc *MessageCollector) collect(m Message) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.messages = append(mc.messages, m)
}
