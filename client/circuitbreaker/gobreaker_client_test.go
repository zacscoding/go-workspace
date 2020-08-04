package circuitbreaker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sony/gobreaker"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestGoBreaker(t *testing.T) {
	// start stub server
	e := NewStubServer()
	go func() {
		if err := e.Run(":3000"); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)
	st := gobreaker.Settings{
		Name:        "service1",
		MaxRequests: 0,
		Interval:    0,
		Timeout:     2 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			b, _ := json.Marshal(counts)
			log.Printf("ReadyToTrip: %s", string(b))
			rate := float32(counts.ConsecutiveFailures) / float32(counts.Requests)
			if rate >= 0.5 {
				return true
			}
			return false
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit[%s] %v -> %v", name, from, to)
		},
	}
	cb := gobreaker.NewCircuitBreaker(st)

	call(cb, "1", "0", true)  // success
	call(cb, "2", "3", false) // fail -> circuit open
	call(cb, "3", "3", false) // fail -> fast fail
	time.Sleep(2 * time.Second)
	call(cb, "4", "0", true)  // half open -> success
	call(cb, "5", "2", false) // half open -> success
}

func call(cb *gobreaker.CircuitBreaker, id, sleep string, success bool) {
	url := fmt.Sprintf("http://localhost:3000/sleep/%s/%s/%v", id, sleep, success)
	log.Println("--------------------------------------------------")
	log.Println("Try ::", id, " =>", url)

	body, err := cb.Execute(func() (interface{}, error) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode >= 300 {
			return nil, errors.New(string(body))
		}
		return body, nil
	})
	if err != nil {
		log.Println("error :", err.Error())
		return
	}
	log.Println("success :", string(body.([]byte)))
}
