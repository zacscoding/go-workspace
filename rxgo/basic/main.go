package main

import (
	"context"
	"fmt"
	"github.com/reactivex/rxgo/v2"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const debug = true

type StringItem struct {
	Val string
}

func printf(format string, v ...interface{}) {
	prefix := fmt.Sprintf("[Goroutine-%d] ", goroutineID())
	log.Printf(prefix+format, v...)
}

func main() {
	printf("Start to tests rxgo")
	ch := make(chan rxgo.Item)
	go produceItems(ch)

	observable := rxgo.FromChannel(ch)
	observable = observable.
		// skip to process if value starts with "skip"
		Filter(func(item interface{}) bool {
			i := item.(*StringItem)
			if debug {
				printf("[Filter] val: %s, return: %v", i.Val, !strings.HasPrefix("skip", i.Val))
			}
			return !strings.HasPrefix(i.Val, "skip")
		}).
		// convert to uppercase given string value with 10 pools and 1 buffered channel
		Map(processUppercase,
			rxgo.WithPool(10),
			rxgo.WithBufferedChannel(1)).
		// replcae all "_" to "-" with default options
		Map(processReplaceAll)

	for item := range observable.Observe() {
		if item.Error() {
			printf("[Result] item: %v, err: %v", item.V, item.E)
		} else {
			printf("[Result] item: %v", item.V)
		}
	}
	// Output:
	//2021/05/26 21:26:19 [Goroutine-1] Start to tests rxgo
	//2021/05/26 21:26:19 [Goroutine-20] [Filter] val: UPPER_CASE_VALUE_1, return: true
	//2021/05/26 21:26:19 [Goroutine-26] [Map-1] val: UPPER_CASE_VALUE_1
	//2021/05/26 21:26:19 [Goroutine-32] [Map-2] val: UPPER_CASE_VALUE_1
	//2021/05/26 21:26:19 [Goroutine-1] [Result] item: &{UPPER-CASE-VALUE-1}
	//2021/05/26 21:26:21 [Goroutine-20] [Filter] val: skip_value_2, return: true
	//2021/05/26 21:26:21 [Goroutine-20] [Filter] val: lower_case_value_0, return: true
	//2021/05/26 21:26:21 [Goroutine-27] [Map-1] val: lower_case_value_0
	//2021/05/26 21:26:21 [Goroutine-32] [Map-2] val: lower_case_value_0
	//2021/05/26 21:26:21 [Goroutine-1] [Result] item: &{lower-case-value-0}
	//2021/05/26 21:26:21 [Goroutine-20] [Filter] val: error_value_3, return: true
	//2021/05/26 21:26:21 [Goroutine-28] [Map-1] val: error_value_3
	//2021/05/26 21:26:21 [Goroutine-1] [Result] item: <nil>, err: force error. val:error_value_3
}

func processUppercase(_ context.Context, item interface{}) (interface{}, error) {
	i := item.(*StringItem)
	if debug {
		printf("[Map-1] val: %s", i.Val)
	}
	if strings.HasPrefix(i.Val, "error") {
		return i, fmt.Errorf("force error. val:%s", i.Val)
	}
	return i, nil
}

func processReplaceAll(_ context.Context, item interface{}) (interface{}, error) {
	i := item.(*StringItem)
	if debug {
		printf("[Map-2] val: %s", i.Val)
	}
	i.Val = strings.ReplaceAll(i.Val, "_", "-")
	return i, nil
}

func produceItems(ch chan rxgo.Item) {
	wg := sync.WaitGroup{}
	for i := 0; i < 4; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
			var val string
			switch i % 4 {
			case 0:
				val = fmt.Sprintf("lower_case_value_%d", i)
			case 1:
				val = fmt.Sprintf("UPPER_CASE_VALUE_%d", i)
			case 2:
				val = fmt.Sprintf("skip_value_%d", i)
			case 3:
				val = fmt.Sprintf("error_value_%d", i)
			}
			ch <- rxgo.Item{
				V: &StringItem{
					Val: val,
				},
			}
		}()
	}
	wg.Wait()
	close(ch)
}

func goroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
