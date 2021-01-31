package fastclientconfigs

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMaxConns(t *testing.T) {
	e := gin.New()
	e.GET("sleep", func(ctx *gin.Context) {
		timeValue, ok := ctx.GetQuery("time")
		if ok {
			timeSecs, err := strconv.Atoi(timeValue)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
				return
			}
			time.Sleep(time.Duration(timeSecs) * time.Millisecond)
		}
		ctx.Status(http.StatusOK)
	})
	go func() {
		if err := e.Run(":3000"); err != nil {
			log.Fatal(err)
		}
	}()

	cases := []struct {
		Name               string
		MaxConnsPerHost    int
		MaxConnWaitTimeout time.Duration
		Workers            int
		SleepMills         int
		AssertFunc         func(t *testing.T, success, failure int32)
	}{
		{
			Name:               "Failure 5 workers",
			MaxConnsPerHost:    10,
			MaxConnWaitTimeout: time.Second,
			Workers:            15,
			SleepMills:         2000,
			AssertFunc: func(t *testing.T, success, failure int32) {
				assert.EqualValues(t, 10, success)
				assert.EqualValues(t, 5, failure)
			},
		}, {
			Name:               "Success all",
			MaxConnsPerHost:    10,
			MaxConnWaitTimeout: time.Second * 3,
			Workers:            15,
			SleepMills:         2000,
			AssertFunc: func(t *testing.T, success, failure int32) {
				assert.EqualValues(t, 15, success)
				assert.EqualValues(t, 0, failure)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			cli := fasthttp.Client{
				MaxConnsPerHost:               tc.MaxConnsPerHost,
				//MaxIdleConnDuration:           0,
				//MaxConnDuration:               0,
				//MaxIdemponentCallAttempts:     0,
				//ReadTimeout:                   0,
				//WriteTimeout:                  0,
				MaxConnWaitTimeout:            tc.MaxConnWaitTimeout,
			}
			var (
				wg      = sync.WaitGroup{}
				success = int32(0)
				failure = int32(0)
			)
			for i := 1; i <= tc.Workers; i++ {
				wg.Add(1)
				id := fmt.Sprintf("Client-%d", i)
				go func() {
					defer wg.Done()
					log.Printf("[#%s] Try to request", id)
					req := fasthttp.AcquireRequest()
					resp := fasthttp.AcquireResponse()
					defer fasthttp.ReleaseRequest(req)
					defer fasthttp.ReleaseResponse(resp)
					req.SetRequestURI(fmt.Sprintf("http://localhost:3000/sleep?time=%d", tc.SleepMills))
					if err := cli.Do(req, resp); err == nil {
						atomic.AddInt32(&success, 1)
					} else {
						atomic.AddInt32(&failure, 1)
						log.Printf("[#%s] Err:%s", id, err.Error())
					}
				}()
			}
			wg.Wait()
			log.Printf(">> Success:%d, Failure:%d", success, failure)
			tc.AssertFunc(t, success, failure)
		})
	}
}
