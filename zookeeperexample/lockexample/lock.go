package main

import (
	"errors"
	"fmt"
	"github.com/go-zookeeper/zk"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrDeadlock is returned by Lock when trying to lock twice without unlocking first
	ErrDeadlock = errors.New("zk: trying to acquire a lock twice")
	// ErrTimeoutAcquireLock is returned by Lock when occur timeout.
	ErrTimeoutAcquireLock = errors.New("zk: timeout to acquire a lock")
	// ErrNotLocked is returned by Unlock when trying to release a lock that has not first be acquired.
	ErrNotLocked = errors.New("zk: not locked")
)

// Lock is a mutual exclusion lock.
type Lock struct {
	c        *zk.Conn
	path     string
	acl      []zk.ACL
	lockPath string
	seq      int
}

// NewLock creates a new lock instance using the provided connection, path, and acl.
// The path must be a node that is only used by this lock. A lock instances starts
// unlocked until Lock() is called.
func NewLock(c *zk.Conn, path string, acl []zk.ACL) *Lock {
	return &Lock{
		c:    c,
		path: path,
		acl:  acl,
	}
}

func parseSeq(path string) (int, error) {
	parts := strings.Split(path, "-")
	// python client uses a __LOCK__ prefix
	if len(parts) == 1 {
		parts = strings.Split(path, "__")
	}
	return strconv.Atoi(parts[len(parts)-1])
}

// Lock attempts to acquire the lock. It works like LockWithData, but it doesn't
// write any data to the lock node.
func (l *Lock) Lock(ttl time.Duration) error {
	return l.LockWithData([]byte{}, ttl)
}

// LockWithData attempts to acquire the lock, writing data into the lock node.
// It will wait to return until the lock is acquired or an error occurs. If
// this instance already has the lock then ErrDeadlock is returned.
func (l *Lock) LockWithData(data []byte, ttl time.Duration) error {
	if l.lockPath != "" {
		return ErrDeadlock
	}

	prefix := fmt.Sprintf("%s/lock-", l.path)

	path := ""
	var err error
	for i := 0; i < 3; i++ {
		path, err = l.c.CreateProtectedEphemeralSequential(prefix, data, l.acl)
		if err == zk.ErrNoNode {
			// Create parent node.
			parts := strings.Split(l.path, "/")
			pth := ""
			for _, p := range parts[1:] {
				var exists bool
				pth += "/" + p
				exists, _, err = l.c.Exists(pth)
				if err != nil {
					return err
				}
				if exists == true {
					continue
				}
				_, err = l.c.Create(pth, []byte{}, 0, l.acl)
				if err != nil && err != zk.ErrNodeExists {
					return err
				}
			}
		} else if err == nil {
			break
		} else {
			return err
		}
	}
	if err != nil {
		return err
	}

	seq, err := parseSeq(path)
	if err != nil {
		return err
	}

	timer := time.NewTimer(ttl)
	defer timer.Stop()
	for {
		children, _, err := l.c.Children(l.path)
		if err != nil {
			return err
		}

		lowestSeq := seq
		prevSeq := -1
		prevSeqPath := ""
		for _, p := range children {
			s, err := parseSeq(p)
			if err != nil {
				return err
			}
			if s < lowestSeq {
				lowestSeq = s
			}
			if s < seq && s > prevSeq {
				prevSeq = s
				prevSeqPath = p
			}
		}

		if seq == lowestSeq {
			log.Printf("[Lock] success to acquire lock-%s", l.path)
			// Acquired the lock
			break
		}

		// Wait on the node next in line for the lock
		log.Println("[Lock] Wait on the node next in line for the lock")
		_, _, ch, err := l.c.GetW(l.path + "/" + prevSeqPath)
		if err != nil && err != zk.ErrNoNode {
			return err
		} else if err != nil && err == zk.ErrNoNode {
			// try again
			continue
		}

		select {
		case ev := <-ch:
			log.Println("[Lock] event:", ev)
			if ev.Err != nil {
				return ev.Err
			}
		case <-timer.C:
			log.Println("[Lock] timeout")
			return ErrTimeoutAcquireLock
		}
	}

	l.seq = seq
	l.lockPath = path
	return nil
}

// Unlock releases an acquired lock. If the lock is not currently acquired by
// this Lock instance than ErrNotLocked is returned.
func (l *Lock) Unlock() error {
	if l.lockPath == "" {
		return ErrNotLocked
	}
	if err := l.c.Delete(l.lockPath, -1); err != nil {
		return err
	}
	l.lockPath = ""
	l.seq = 0
	return nil
}
