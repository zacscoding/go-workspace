package fetchque

import (
	"fmt"
	"github.com/asaskevich/EventBus"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type FetcherQueue1 struct {
	EventBus EventBus.Bus
	Keys     map[string]struct{}
	KeyMutex sync.Mutex
}

func (f *FetcherQueue1) FetchData1(key string, callback Data1Callback) error {
	f.KeyMutex.Lock()
	topic := fmt.Sprintf("data1:%s", key)
	f.Keys[topic] = struct{}{}
	f.KeyMutex.Unlock()
	//onceFlag := int32(0)
	//wrapperCallback := func(data *Data1) {
	//	if atomic.CompareAndSwapInt32(&onceFlag, 0, 1) {
	//		callback(data)
	//	}
	//}
	return f.EventBus.SubscribeOnceAsync(topic, callback)
}

func (f *FetcherQueue1) FetchData2(key string, callback Data2Callback) error {
	f.KeyMutex.Lock()
	topic := fmt.Sprintf("data2:%s", key)
	f.Keys[topic] = struct{}{}
	f.KeyMutex.Unlock()

	//onceFlag := int32(0)
	//wrapperCallback := func(data *Data2) {
	//	if atomic.CompareAndSwapInt32(&onceFlag, 0, 1) {
	//		callback(data)
	//	}
	//}
	return f.EventBus.SubscribeOnceAsync(topic, callback)
}

func NewFetchQueue1(interval time.Duration) *FetcherQueue1 {
	fetchQueue := FetcherQueue1{
		EventBus: EventBus.New(),
		Keys:     make(map[string]struct{}),
	}
	go fetchQueue.loopFetch(interval)
	return &fetchQueue
}

func (f *FetcherQueue1) loopFetch(interval time.Duration) {
	var (
		count  = 0
		ticker = time.NewTicker(interval)
	)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			count++
			f.KeyMutex.Lock()
			var topics []string
			for k := range f.Keys {
				topics = append(topics, k)
				delete(f.Keys, k)
			}
			f.KeyMutex.Unlock()

			if len(topics) == 0 {
				log.Println("Skip to fetch data because of empty topics")
				continue
			}
			for _, topic := range topics {
				splitted := strings.Split(topic, ":")
				if len(splitted) != 2 {
					return
				}
				var (
					dtype = splitted[0]
					key   = splitted[1]
				)
				log.Println("Try to fetch ", dtype, ">", key)
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				value := fmt.Sprintf("data-%d", count)
				if dtype == "data1" {
					f.EventBus.Publish(topic, Data1{
						RequestedKey: topic,
						Value:        value,
					})
				} else {
					f.EventBus.Publish(topic, Data2{
						RequestedKey: topic,
						Value:        value,
					})
				}
				//go func(topic string) {
				//	splitted := strings.Split(topic, ":")
				//	if len(splitted) != 2 {
				//		return
				//	}
				//	var (
				//		dtype = splitted[0]
				//		key   = splitted[1]
				//	)
				//	log.Println("Try to fetch ", dtype, ">", key)
				//	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				//	value := fmt.Sprintf("data-%d", count)
				//	if dtype == "data1" {
				//		f.EventBus.Publish(topic, Data1{
				//			RequestedKey: topic,
				//			Value:        value,
				//		})
				//	} else {
				//		f.EventBus.Publish(topic, Data2{
				//			RequestedKey: topic,
				//			Value:        value,
				//		})
				//	}
				//}(topic)
			}
			f.EventBus.WaitAsync()
		}
	}
}
