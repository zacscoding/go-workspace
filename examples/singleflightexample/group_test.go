package singleflightexample

import (
	"golang.org/x/sync/singleflight"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type LocalCache struct {
	mu    sync.Mutex
	cache map[string]string

	cacheHit   int
	cacheTotal int
	setCalled  int
}

func (c *LocalCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheTotal++

	v, ok := c.cache[key]
	if ok {
		c.cacheHit++
	}
	return v, ok
}

func (c *LocalCache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.setCalled++
	c.cache[key] = value
}

type CacheItem struct {
	Key   string
	Value func(i *CacheItem) string
}

func TestSingleFlightGroup(t *testing.T) {
	localCache := &LocalCache{
		cache: make(map[string]string),
	}
	key1 := "key1"
	key2 := "key2"
	keys := []string{key1, key2}

	group := singleflight.Group{}
	wg := sync.WaitGroup{}
	completed := int32(0)
	valueCalled := int32(0)
	sharedCount := int32(0)

	for i := 0; i < 50; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			item := CacheItem{
				Key: keys[i%len(keys)],
				Value: func(i *CacheItem) string {
					atomic.AddInt32(&valueCalled, 1)
					time.Sleep(time.Duration(rand.Intn(100)))
					switch i.Key {
					case key1:
						return "value1"
					case key2:
						return "value2"
					default:
						return "unknown"
					}
				},
			}
			_, _, shared := group.Do(item.Key, func() (interface{}, error) {
				v, ok := localCache.Get(item.Key)
				if ok {
					return v, nil
				}
				v = item.Value(&item)
				localCache.Set(item.Key, v)
				return v, nil
			})
			//log.Printf("Task-%d > res: %s, err: %v, shared: %v", i, res, err, shared)
			atomic.AddInt32(&completed, 1)
			if shared {
				atomic.AddInt32(&sharedCount, 1)
			}
		}()
	}
	wg.Wait()

	log.Printf("Cache hit: %d/%d", localCache.cacheHit, localCache.cacheTotal)
	log.Println("Set called count:", localCache.setCalled)
	log.Println("Completed count:", completed)
	log.Println("Value called count", valueCalled)
	log.Println("Shared count:", sharedCount)
}
