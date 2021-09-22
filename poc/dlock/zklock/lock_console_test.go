package zklock

import (
	"context"
	"github.com/go-zookeeper/zk"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go-workspace/poc/dlock"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestUnlocks(t *testing.T) {
	conn := newZKConn()
	defer conn.Close()

	registry := NewZKLockRegistry(conn, WithDefaultTTL(time.Second*10), WithLogger(&dlock.StdLogger{}))
	lock := registry.NewLock("task1")

	// Try to acquire a lock => Success
	err := lock.Lock(context.Background(), time.Second*3)
	assert.NoError(t, err)

	// Sleep more than ttl
	time.Sleep(time.Second * 4)

	// Unlock => Fail
	err = lock.Unlock()
	assert.Equal(t, dlock.ErrNotLocked, err)

	// Try to acquire a lock => Success
	err = lock.Lock(context.Background(), time.Second*5)

	// Sleep less than ttl
	time.Sleep(time.Second * 3)

	// Unlock => Success
	err = lock.Unlock()
	assert.NoError(t, err)
}

func TestExtend(t *testing.T) {
	m := sync.Mutex{}
	m.Lock()

	conn := newZKConn()
	defer conn.Close()

	registry := NewZKLockRegistry(conn, WithDefaultTTL(time.Second*10), WithLogger(&dlock.StdLogger{}))
	lock := registry.NewLock("task1")

	err := lock.Lock(context.Background(), time.Second*3)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn := newZKConn()
		defer conn.Close()

		registry := NewZKLockRegistry(conn, WithDefaultTTL(time.Second*10), WithLogger(&dlock.StdLogger{}))
		lock := registry.NewLock("task1")

		for i := 0; i < 10; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := lock.Lock(ctx, time.Second*5)
			log.Println("[Lock-2] try to acquire lock. err:", err)
			cancel()
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ok, err := lock.Extend()
				if !ok || err != nil {
					log.Printf("[Lock-1] failed to extend lock. ok: %v, err: %v", ok, err)
				}
			}
		}
	}()

	wg.Wait()
}

func TestCheckNodes(t *testing.T) {
	conn := newZKConn()
	defer conn.Close()

	path := "/task1"
	children, _, err := conn.Children(path)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Check %s => %d", path, len(children))
	for _, c := range children {
		log.Println(">", c)
	}
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
