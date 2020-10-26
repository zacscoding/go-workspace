package pipeline

import (
	"fmt"
	"testing"
)

var (
	generator = func(done <-chan interface{}, values ...int) <-chan int {
		stream := make(chan int, len(values))
		go func() {
			defer close(stream)
			for _, value := range values {
				select {
				case <-done:
					return
				case stream <- value:
				}
			}
		}()
		return stream
	}

	multiply = func(done <-chan interface{}, stream <-chan int, multiplier int) <-chan int {
		multipliedStream := make(chan int)
		go func() {
			defer close(multipliedStream)
			for i := range stream {
				select {
				case <-done:
					return
				case multipliedStream <- i * multiplier:
				}
			}

		}()
		return multipliedStream
	}

	add = func(done <-chan interface{}, stream <-chan int, additive int) <-chan int {
		addedStream := make(chan int)
		go func() {
			defer close(addedStream)
			for i := range stream {
				select {
				case <-done:
					return
				case addedStream <- i + additive:
				}
			}
		}()
		return addedStream
	}
)

func TestPipeline(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	intStream := generator(done, 1, 2, 3, 4)
	// 3 * ( (i * 2) + 1 )
	pipeline := multiply(done, add(done, multiply(done, intStream, 2), 1), 3)

	for v := range pipeline {
		fmt.Println(v)
	}
	// Output
	//9
	//15
	//21
	//27
}
