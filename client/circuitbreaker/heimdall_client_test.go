package circuitbreaker

import (
	"fmt"
	"github.com/gojektech/heimdall/v6/hystrix"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestHeimdall(t *testing.T) {
	// start stub server
	e := NewStubServer()
	go func() {
		if err := e.Run(":3000"); err != nil {
			log.Fatal(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// Create a new hystrix-wrapped HTTP client with the command name, along with other required options
	client := hystrix.NewClient(
		hystrix.WithHTTPTimeout(10*time.Millisecond),
		hystrix.WithCommandName("service1"),
		hystrix.WithHystrixTimeout(1*time.Second),
		hystrix.WithMaxConcurrentRequests(30),
		hystrix.WithErrorPercentThreshold(90),
	)

	for i := 0; i < 10; i++ {
		var (
			success bool
			sleep   int
		)
		if i == 0 {
			success = true
		} else {
			success = false
			sleep = 2
		}

		url := fmt.Sprintf("http://localhost:3000/sleep/%d/%d/%v", i, sleep, success)
		log.Println("--------------------------------------------------")
		log.Println("Try ::", i, " =>", url)

		res, err := client.Get(url, nil)
		if err != nil {
			log.Println("err:", err.Error())
			continue
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Read body err:", err.Error())
		}
		log.Println("Status:", res.Status, ", Body:", string(body))
		time.Sleep(500 * time.Millisecond)
	}
}
