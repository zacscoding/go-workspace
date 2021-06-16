package main

import (
	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

var (
	zkServers = []string{
		"localhost:2181",
	}
	rootPath = "/session-tests"

	stateNames = map[zk.State]string{
		zk.StateUnknown:           "StateUnknown",
		zk.StateDisconnected:      "StateDisconnected",
		zk.StateConnectedReadOnly: "StateConnectedReadOnly",
		zk.StateSaslAuthenticated: "StateSaslAuthenticated",
		zk.StateExpired:           "StateExpired",
		zk.StateAuthFailed:        "StateAuthFailed",
		zk.StateConnecting:        "StateConnecting",
		zk.StateConnected:         "StateConnected",
		zk.StateHasSession:        "StateHasSession",
	}
)

func main() {
	c, eventCh, err := zk.Connect(zkServers, time.Second*3)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	c.SetLogger(&zkLogger{})
	if _, err := c.Create(rootPath, []byte{}, zk.FlagSequence, zk.WorldACL(zk.PermAll)); err != nil {
		log.Fatal(err)
	}

	go func() {
		for event := range eventCh {
			if event.Type == zk.EventSession {
				log.Printf("[EventLoop] session state: %s", stateNames[event.State])
				switch event.State {
				case zk.StateExpired:
					log.Println("[EventLoop] session expired..")
				case zk.StateConnected:
					log.Println("[EventLoop] session started..")
				}
			}
		}
	}()

	done := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				path := rootPath + "/" + uuid.New().String()
				if _, err := c.Create(path, []byte("temporary"), zk.FlagEphemeral, zk.WorldACL(zk.PermAll)); err != nil {
					log.Printf("failed to create a node %s. err: %v", path, err)
				}
			case <-done:
				return
			}
		}
	}()
	time.Sleep(10 * time.Second)
	close(done)

	log.Println("Check children")
	children, _, err := c.Children(rootPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("> Len:", len(children))
	for _, c := range children {
		log.Println("> Children:", c)
	}
}

func createRecursive(conn *zk.Conn, path string) error {
	if path == "/" {
		return nil
	}

	parts := strings.Split(path, "/")
	for i := 2; i <= len(parts); i++ {
		p := strings.Join(parts[:i], "/")
		// If the rootpath exists, skip the Create process to avoid "zk: not authenticated" error
		exist, _, errExists := conn.Exists(p)
		if !exist {
			_, err := conn.Create(p, []byte{}, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			// Ignore when the node exists already
			if (err != nil) && (err != zk.ErrNodeExists) {
				return err
			}
		} else {
			return errExists
		}
	}
	return nil
}

type zkLogger struct {
}

func (l *zkLogger) Printf(foramt string, v ...interface{}) {
	log.Printf("[Zookeeper]"+foramt, v)
}
