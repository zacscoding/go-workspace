package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type collector struct {
	messages []string
	mutex    sync.Mutex
}

func (c *collector) collect(m string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.messages = append(c.messages, m)
}

func main() {
	c := &collector{}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		p := &producer{
			name: fmt.Sprintf("Worker-%d", i+1),
			c:    c,
		}
		go p.start(&wg)
	}
	wg.Wait()
	fmt.Println(len(c.messages))
}

type producer struct {
	name string
	c    *collector
}

func (p *producer) start(wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 0; i < 10; i++ {
		p.c.collect(fmt.Sprintf("%s-%d", p.name, i))
		time.Sleep(time.Millisecond * 200)
	}
	log.Printf("Terminate-%s", p.name)
}
