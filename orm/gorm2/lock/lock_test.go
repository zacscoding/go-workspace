// create table and stored procedures at ./ddl.sql before executing tests,
package lock

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAcquireLock(t *testing.T) {
	var (
		db       = NewDB(t)
		l        = &Lock{db: db}
		taskName = "my-task-1"
		ttl      = 5 * time.Second
	)

	unlock, err := l.Lock(context.Background(), taskName, ttl)
	assert.NoError(t, err)
	t.Logf("lock-1 acquire err: %v", err)

	_, err = l.Lock(context.Background(), taskName, ttl)
	assert.Error(t, err)
	t.Logf("lock-2 acquire err: %v", err)

	err = unlock()
	assert.NoError(t, err)
	t.Logf("release lock-1 err: %v", err)

	_, err = l.Lock(context.Background(), taskName, ttl)
	assert.NoError(t, err)
	t.Logf("lock-3 acquire lock err: %v", err)
	// Output:
	// lock_test.go:28: lock-1 acquire err: <nil>
	// lock_test.go:32: lock-2 acquire err: lock already in use
	// lock_test.go:36: release lock-1 err: <nil>
	// lock_test.go:40: lock-3 acquire lock err: <nil>
}

func TestAcquireLockConcurrency(t *testing.T) {
	var (
		db       = NewDB(t)
		l        = &Lock{db: db}
		taskName = "my-task-1"
		ttl      = 5 * time.Second
		success  = int32(0)
		wg       = sync.WaitGroup{}
	)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
			_, err := l.Lock(context.TODO(), taskName, ttl)
			if err == nil {
				atomic.AddInt32(&success, 1)
			} else {
				//printf("failed to acquire lock. err: %v", err)
			}
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 1, success)
	t.Logf("Success: %d", success)
}

func NewDB(t *testing.T) *gorm.DB {
	dsn := "root:password@tcp(127.0.0.1:13306)/my_database?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Silent, // Log level
				Colorful:      true,
			},
		),
	})
	if err != nil || db == nil {
		t.Fatalf("open db:%v", err)
	}
	return db
}
