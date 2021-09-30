package functions

import (
	"github.com/go-zookeeper/zk"
	"log"
	"sync"
	"testing"
	"time"
)

const enableZKLog = true

func TestSequantialNode(t *testing.T) {
	conn, err := newZKConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// 1) create "/app1"
	app1Path, err := conn.Create("/app1", []byte{}, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		conn.Delete(app1Path, -1)
	}()
	log.Printf("Success to create a /app1 node > path: %s", app1Path)

	// 2) watch
	_, _, events, err := conn.GetW(app1Path)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			select {
			case e := <-events:
				if e.Err != nil {
					log.Printf("[Watcher] error occur. err: %v", e.Err)
					return
				}
				log.Printf("[Watcher] event occur. Type: %d, State: %d, Path: %s", e.Type, e.State, e.Path)
				if e.Type == zk.EventNodeDeleted {
					return
				}
			}
		}
	}()

	// 3) create sequantial
	server1Path, err := conn.Create("/app1/server", []byte{}, zk.FlagEphemeral|zk.FlagSequence, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		conn.Delete(server1Path, -1)
	}()
	log.Printf("Success to create a /app1/server node > path: %s", server1Path)
	// Output
	//2021/09/27 13:51:55 Success to create a /app1 node > path: /app1
	//2021/09/27 13:51:55 Success to create a /app1/server node > path: /app1/server0000000000
	//2021/09/27 13:51:55 [Watcher] event occur. Type: 2, State: 3, Path: /app1
}

func newZKConnection() (*zk.Conn, error) {
	conn, ch, err := zk.Connect([]string{"localhost:2181"}, time.Minute, zk.WithLogger(&zkLogger{}))
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				switch e.State {
				case zk.StateConnected, zk.StateHasSession:
					return
				}
			}
		}
	}()
	wg.Wait()
	return conn, nil
}

type zkLogger struct{}

func (l *zkLogger) Printf(foramt string, v ...interface{}) {
	if true {
		log.Printf("[Zookeeper]"+foramt, v)
	}
}
