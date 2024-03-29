package zklock

import (
	"context"
	"fmt"
	"github.com/go-zookeeper/zk"
	"go-workspace/poc/dlock"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	flagPersistent           = 0
	flagPersistentSequential = 2
	flagEphemeral            = 1
	flagEphemeralSequential  = 3
)

const (
	stateNoLock = 0
	stateLock   = 1
)

const (
	maxAttempts = 3
)

type Option func(r *zkLockRegistry)

func WithACLs(acls []zk.ACL) Option {
	return func(r *zkLockRegistry) {
		r.acls = acls
	}
}

func WithLogger(l dlock.Logger) Option {
	return func(r *zkLockRegistry) {
		r.logger = l
	}
}

func WithDefaultTTL(ttl time.Duration) Option {
	return func(r *zkLockRegistry) {
		r.defaultTTL = ttl
	}
}

type zkLockRegistry struct {
	conn       *zk.Conn
	acls       []zk.ACL
	logger     dlock.Logger
	defaultTTL time.Duration
}

func NewZKLockRegistry(conn *zk.Conn, opts ...Option) dlock.LockRegistry {
	registry := zkLockRegistry{
		conn:       conn,
		acls:       zk.WorldACL(zk.PermAll),
		logger:     &dlock.NoopLogger{},
		defaultTTL: dlock.DefaultLockTTL,
	}
	for _, opt := range opts {
		opt(&registry)
	}
	return &registry
}

func (z *zkLockRegistry) NewLock(key string) dlock.Lock {
	if key[0] != '/' {
		key = "/" + key
	}
	return &ZKLock{
		conn:   z.conn,
		path:   key,
		acls:   zk.WorldACL(zk.PermAll),
		logger: z.logger,
		ttl:    z.defaultTTL,
	}
}

type ZKLock struct {
	conn   *zk.Conn
	path   string
	acls   []zk.ACL
	logger dlock.Logger
	ttl    time.Duration

	acquired int32
	// lockPath is setted after acquiring a lock.
	lockPath    string
	seq         int
	unlockTimer *time.Timer
	timerStopCH chan struct{}
}

func (zl *ZKLock) Lock(ctx context.Context, ttl time.Duration) error {
	return zl.LockWithData(ctx, ttl, []byte{})
}

func (zl *ZKLock) LockWithData(ctx context.Context, ttl time.Duration, data []byte) error {
	// client tried to acquire lock twice
	if zl.isAcquired() {
		return dlock.ErrDeadlock
	}
	if ttl == 0 {
		ttl = zl.ttl
	}

	var (
		pathname = fmt.Sprintf("%s/lock-", zl.path)
		path     string
		err      error
	)

	for i := 0; i < maxAttempts; i++ {
		if isDone(ctx) {
			return dlock.ErrTimeoutAcquireLock
		}
		// ------------------------------------------------------------
		// Create ephemeral sequential node
		// ------------------------------------------------------------
		// Path will be "{prefix}/{guid}-lock-{seq}" if success to create a node.
		// e.g: "/myproject/service1/task1/_c_35e5b99041f2c7203d1a0a1c18a1e609-lock-0000000000"
		path, err = zl.conn.CreateProtectedEphemeralSequential(pathname, data, zl.acls)
		if err == nil {
			break
		}
		if err != zk.ErrNoNode {
			return err
		}
		// ------------------------------------------------------------
		// Create recursive
		// ------------------------------------------------------------
		parts := strings.Split(zl.path, "/")
		for i := 2; i <= len(parts); i++ {
			part := strings.Join(parts[:i], "/")
			exists, _, err := zl.conn.Exists(part)
			if exists {
				continue
			}
			if err != nil {
				return err
			}
			_, err = zl.conn.Create(part, []byte{}, flagPersistent, zl.acls)
			// Ignore when the node exists already
			if err != nil && err != zk.ErrNodeExists {
				return err
			}
		}
	}

	seq, err := parseSequence(path)
	if err != nil {
		return err
	}

	for {
		if isDone(ctx) {
			return dlock.ErrTimeoutAcquireLock
		}
		// ------------------------------------------------------------
		// Getting children to check acquiring lock.
		// If current sequence is equals to lowestSeq, then success,
		// Otherwise watch a node which less than current sequence.
		// ------------------------------------------------------------
		children, _, err := zl.conn.Children(zl.path)
		if err != nil {
			return err
		}

		var (
			lowestSeq   = seq
			prevSeq     = -1
			prevSeqPath string
		)

		for _, c := range children {
			s, err := parseSequence(c)
			if err != nil {
				return err
			}
			// find the lowest sequence.
			if s < lowestSeq {
				lowestSeq = s
			}
			// find previous sequence.
			if s < seq && s > prevSeq {
				prevSeq = s
				prevSeqPath = c
			}
		}

		// If current sequence is the lowest, then success to acquire a lock.
		if seq == lowestSeq {
			break
		}

		_, _, ch, err := zl.conn.GetW(zl.path + "/" + prevSeqPath)
		if err != nil {
			if err != zk.ErrNoNode {
				return err
			}
			// try again because previous path was deleted.
			continue
		}

		select {
		case e := <-ch:
			if e.Err != nil {
				return e.Err
			}
		case <-ctx.Done():
			zl.conn.Delete(path, -1)
			return dlock.ErrTimeoutAcquireLock
		}
	}

	// check context is done or not
	if isDone(ctx) {
		zl.unlock()
		return dlock.ErrTimeoutAcquireLock
	}

	atomic.CompareAndSwapInt32(&zl.acquired, stateNoLock, stateLock)
	zl.seq = seq
	zl.lockPath = path
	zl.unlockTimer = time.NewTimer(ttl)
	zl.timerStopCH = make(chan struct{}, 1)
	go func() {
		select {
		case <-zl.unlockTimer.C:
			zl.unlock()
		case <-zl.timerStopCH:
		}
	}()
	return nil
}

func (zl *ZKLock) Extend() (bool, error) {
	if !zl.isAcquired() {
		return false, dlock.ErrNotLocked
	}
	return zl.unlockTimer.Reset(zl.ttl), nil
}

func (zl *ZKLock) Unlock() error {
	if !zl.isAcquired() {
		return dlock.ErrNotLocked
	}
	zl.unlockTimer.Stop()
	close(zl.timerStopCH)
	return zl.unlock()
}

func (zl *ZKLock) unlock() error {
	if err := zl.conn.Delete(zl.lockPath, -1); err != nil {
		if err == zk.ErrNoNode {
			zl.clearLockContexts()
			return dlock.ErrNotLocked
		}
		return err
	}
	zl.clearLockContexts()
	return nil
}

func (zl *ZKLock) clearLockContexts() {
	if !atomic.CompareAndSwapInt32(&zl.acquired, stateLock, stateNoLock) {
		return
	}
	zl.lockPath = ""
	zl.seq = 0
	zl.unlockTimer = nil
	zl.timerStopCH = nil
}

func (zl *ZKLock) isAcquired() bool {
	return atomic.LoadInt32(&zl.acquired) == stateLock
}

func isDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func parseSequence(path string) (int, error) {
	parts := strings.Split(path, "-")
	// python client uses a __LOCK__ prefix
	if len(parts) == 1 {
		parts = strings.Split(path, "__")
	}
	return strconv.Atoi(parts[len(parts)-1])
}
