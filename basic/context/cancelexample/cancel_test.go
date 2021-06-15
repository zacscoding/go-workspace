package cancelexample

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	w := &Worker{}
	wg := sync.WaitGroup{}
	success := int32(0)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := w.Start()
			if err == nil {
				atomic.AddInt32(&success, 1)
			}
		}()
	}

	wg.Wait()
	assert.EqualValues(t, 1, success)
	assert.True(t, w.Running())

	time.Sleep(time.Second * 5)
	w.Stop()
	time.Sleep(time.Second)
	assert.False(t, w.Running())
}
