package temp

import (
	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	zkServers = []string{
		"localhost:2181",
	}
)

func Test1(t *testing.T) {
	conn, _, err := zk.Connect(zkServers, time.Second)
	assert.NoError(t, err)
	lockPath := "/lock"
	group := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		i := i
		group.Add(1)
		go func() {
			defer group.Done()
			prefix := lockPath + "/uuid-"
			path, err := conn.CreateProtectedEphemeralSequential(prefix, []byte{}, zk.WorldACL(zk.PermAll))
			if err != nil {
				if err == zk.ErrNoNode {
					err = createRecursive(conn, strings.Split(lockPath, "/")[1:])
					assert.NoError(t, err)
				} else {
					t.Fail()
				}
			}
			t.Logf("Worker-%d > %s", i, path)
		}()
	}

	group.Wait()
	conn.Close()
}

func createRecursive(conn *zk.Conn, paths []string) error {
	path := ""
	for _, p := range paths {
		path += "/" + p
		exists, _, err := conn.Exists(path)
		if err != nil {
			return err
		}
		if exists == true {
			continue
		}
		_, err = conn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}
	return nil
}
