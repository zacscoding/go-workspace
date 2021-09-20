package zookeeper

import (
	"fmt"
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDelete(t *testing.T) {
	conn := newZKConn()
	defer conn.Close()

	lockPath := "/myproject/service1/task1"
	path, err := create(conn, lockPath)
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.Delete(path, -1); err != nil {
		log.Printf("First delete err: %v", err)
		return
	}

	if err := conn.Delete(path, -1); err != nil {
		isNoNodeErr := err == zk.ErrNoNode
		log.Printf("Second delete err: %v, isNoNodeErr: %v", err, isNoNodeErr)
		return
	}
}

func TestTemp(t *testing.T) {
	conn1 := newZKConn()
	defer conn1.Close()
	conn2 := newZKConn()
	defer conn2.Close()
	log.Printf("## Conn1 SessionID: %d", conn1.SessionID())
	log.Printf("## Conn2 SessionID: %d", conn2.SessionID())

	lockPath := "/myproject/service1/task1"
	path1, err := create(conn1, lockPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("## Path1 >", path1)
	path2, err := create(conn2, lockPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("## Path2 >", path2)

	checkChildren(conn1, lockPath)

	deleteRecursive(conn1, path1)
	deleteRecursive(conn2, path2)
	log.Println("## Success to delete")
	log.Printf("## Conn1 SessionID: %d", conn1.SessionID())
	log.Printf("## Conn2 SessionID: %d", conn2.SessionID())
}

func checkChildren(conn *zk.Conn, lockPath string) {
	log.Printf("## Check children: %s", lockPath)
	children, _, err := conn.Children(lockPath)
	if err != nil {
		log.Println("## Failed to check children:", err)
		return
	}

	for i, c := range children {
		seq, err := parseSequence(c)
		log.Printf("> path: %s, index: %d, seq: %d, err: %v", c, i, seq, err)
	}
}

func create(conn *zk.Conn, lockPath string) (string, error) {
	var (
		prefix = fmt.Sprintf("%s/lock-", lockPath)
		data   []byte
		acls   = zk.WorldACL(zk.PermAll)
		path   = ""
		err    error
	)

	log.Println("## Create lock path:", lockPath)
	for i := 0; i < 3; i++ {
		log.Printf("> Try to CreateProtectedEphemeralSequential. attempts: %d", i+1)
		// Path will be "{prefix}/{guid}-lock-{seq}" if success to create a node.
		// e.g: "/myproject/service1/task1/_c_35e5b99041f2c7203d1a0a1c18a1e609-lock-0000000000"
		path, err = conn.CreateProtectedEphemeralSequential(prefix, data, acls)
		if err == nil {
			log.Printf("> Success to CreateProtectedEphemeralSequential: %s", path)
			break
		}
		if err != zk.ErrNoNode {
			return "", errors.Wrapf(err, "create ephemeralsequaltial node: %s", prefix)
		}
		// Create parent nodes
		log.Println("> ## Try to create parent nodes..:", path)
		parts := strings.Split(lockPath, "/")
		for i := 2; i <= len(parts); i++ {
			part := strings.Join(parts[:i], "/")
			exists, _, err := conn.Exists(part)
			log.Printf(">> ## Check %s -> exists: %v, err: %v", part, exists, err)
			if exists {
				continue
			}
			if err != nil {
				return "", errors.Wrapf(err, "exists: %s", part)
			}
			_, err = conn.Create(part, []byte{}, flagPersistent, acls)
			// Ignore when the node exists already
			if err != nil && err != zk.ErrNodeExists {
				return "", errors.Wrapf(err, "create: %s", part)
			}
		}
	}
	log.Printf("> ## Success to create path.. %s", path)
	return path, nil
}

func deleteRecursive(conn *zk.Conn, path string) error {
	if path == "/" {
		return nil
	}
	log.Println("## Try to delete:", path)

	parts := strings.Split(path, "/")
	for i := len(parts); i >= 2; i-- {
		part := strings.Join(parts[:i], "/")
		exist, stats, err := conn.Exists(part)
		log.Printf("## Try to delete %s > exists: %v, err: %v", part, exist, err)
		if err != nil {
			return errors.Wrapf(err, "check exists: %s", part)
		}
		if !exist {
			continue
		}
		if err := conn.Delete(part, stats.Version); err != nil {
			return errors.Wrapf(err, "delete path: %s", part)
		}
	}
	return nil
}

func newZKConn() *zk.Conn {
	zkServers := []string{"localhost:2181"}
	c, eventCH, err := zk.Connect(zkServers, time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range eventCH {
			log.Printf("[EventLoop] recv event: %v", event)
			if event.Type == zk.EventSession {
				switch event.State {
				case zk.StateConnected:
					log.Println("[EventLoop] session started..")
				case zk.StateHasSession:
					log.Println("[EventLoop] has session..")
					return
				}
			}
		}
	}()
	wg.Wait()

	c.SetLogger(&zkLogger{})
	return c
}

type zkLogger struct {
}

func (l *zkLogger) Printf(foramt string, v ...interface{}) {
	log.Printf("[ZK] "+foramt, v...)
}
