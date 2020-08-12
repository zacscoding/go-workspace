# Distributed lock  

- [LockRegistry](#LockRegistry)  
- [How to use](#How-to-use)  

# LockRegistry Interface  

```go
type LockRegistry interface {
	// TryLockWithTimeout try to acquire a lock with given taskId and timeout duration.
	// returns a true if success to acquire a lock, otherwise false
	TryLockWithTimeout(taskId string, duration time.Duration) bool

	// TryLockWithContext try to acquire a lock with given taskId and context to cancel.
	// returns a true if success to acquire a lock, otherwise false
	TryLockWithContext(taskId string, ctx context.Context) bool

	// Unlock release a lock with given task id.
	Unlock(taskId string)
}
```  

## Create a new lock registry


```go
// create a new standalone lock i.e single application
sLock := distributedlock.NewStandaloneLockRegistry()

// create a new redis lock
redisCli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
rLock := distributedlock.NewRedisLockRegistry(redisCli, ttl) 
```  

## How to use  

```go
redisCli := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
rLock := distributedlock.NewRedisLockRegistry(redisCli, ttl)

taskId := "shard-0"
if rLock.TryLockWithTimeout(taskId, 1*time.Second) {
    defer rLock.Unlock(taskId)
    // do something
}
```
