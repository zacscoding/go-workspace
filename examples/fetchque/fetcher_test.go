package fetchque

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchQueueConsole(t *testing.T) {
	fetcher := NewFetchQueue1(time.Second * 1)
	counter := int32(0)
	for i := 1; i <= 10; i++ {
		go func(id int) {
			if id > 5 {
				time.Sleep(time.Second * time.Duration(id))
			}
			if id%2 == 0 {
				fetcher.FetchData1("key1", func(data Data1) {
					defer atomic.AddInt32(&counter, 1)
					fmt.Printf("Consumer-%d: Subscribe:%s\n", id, data.String())
				})
			} else {
				fetcher.FetchData2("key2", func(data Data2) {
					defer atomic.AddInt32(&counter, 1)
					fmt.Printf("Consumer-%d: Subscribe:%s\n", id, data.String())
				})
			}
		}(i)
	}
	for {
		if atomic.LoadInt32(&counter) == 10 {
			break
		}
	}
	fmt.Println("Complete!")
}
