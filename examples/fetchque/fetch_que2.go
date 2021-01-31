package fetchque

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"time"
)

type FetcherQueue2 struct {
	activeHandlers  map[string][]reflect.Value
	standbyHandlers map[string][]reflect.Value
	mutex           sync.Mutex
}

func NewFetcherQueue2(interval time.Duration) Fetcher {
	fetcher := FetcherQueue2{
		activeHandlers:  make(map[string][]reflect.Value),
		standbyHandlers: make(map[string][]reflect.Value),
		mutex:           sync.Mutex{},
	}
	return &fetcher
}

func (f *FetcherQueue2) FetchData1(key string, callback Data1Callback) error {
	topic := fmt.Sprintf("data1:%s", key)
	return f.fetchDataInternal(topic, callback)
}

func (f *FetcherQueue2) FetchData2(key string, callback Data2Callback) error {
	topic := fmt.Sprintf("data1:%s", key)
	return f.fetchDataInternal(topic, callback)
}

func (f *FetcherQueue2) loopFetch(interval time.Duration) {
	ticker := time.NewTimer(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 1) getting topics
			copyHandlers := make(map[string][]reflect.Value)
			f.mutex.Lock()
			for topic, handlers := range f.activeHandlers {
				copyHandlers[topic] = make([]reflect.Value, len(handlers))
				copy(copyHandlers[topic], handlers)
			}
			f.mutex.Unlock()
			// 2) fetch data given above topics
			if len(copyHandlers) == 0 {
				log.Println("Skip to fetch data because of empty topics")
				continue
			}
			//wg := sync.WaitGroup{}
			//for _, topic := range topics {
			//	splitted := strings.Split(topic, ":")
			//	if len(splitted) != 2 {
			//		log.Println("Found invalid topic:", topic)
			//		continue
			//	}
			//	var (
			//		dtype = splitted[0]
			//		key   = splitted[1]
			//	)
			//}
		}
	}
}

func (f *FetcherQueue2) fetchDataInternal(topic string, callback interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.activeHandlers[topic] = append(f.activeHandlers[topic], reflect.ValueOf(callback))
	return nil
}
