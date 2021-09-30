package middleware

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

// TestTimeoutMiddleware tests middleware.TimeoutWithConfig
func TestTimeoutMiddleware(t *testing.T) {
	errHandleWait := sync.WaitGroup{}
	errHandleWait.Add(1)
	e := newTimeoutHandler(func(err error, c echo.Context) {
		defer errHandleWait.Done()
		log.Printf("[Server] OnTimeoutRouteErrorHandler is called. err: %v", err)
	})
	go func() {
		if err := e.Start(":8080"); err != nil {
			log.Fatal(err)
		}
	}()

	start := time.Now()
	res, err := http.Get("http://localhost:8080/sleep?sleep=5")
	elapsed := time.Now().Sub(start)
	if err != nil {
		log.Printf("[Client] error occur. err: %v. elapsed: %v", err, elapsed)
		return
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Printf("[Client] StatusCode: %d, Body: %s, Elapsed: %v", res.StatusCode, string(b), elapsed)
	errHandleWait.Wait()
	// Output
	//2021/09/30 23:03:46 ## [Server:handleSleep] sleep 5 secs
	//2021/09/30 23:03:49 [Server] ctx.Request.Context is done
	//2021/09/30 23:03:49 [Client] StatusCode: 503, Body: timeout, Elapsed: 3.003369559s
	//2021/09/30 23:03:51 [Server] Timeout occur. err: http: Handler timeout
}

// TestTimeoutMiddleware tests middleware.TimeoutWithConfig
func TestTimeoutMiddlewareWithServer(t *testing.T) {
	errHandleWait := sync.WaitGroup{}
	errHandleWait.Add(1)
	srv := &http.Server{
		Addr: ":8080",
		Handler: newTimeoutHandler(func(err error, c echo.Context) {
			defer errHandleWait.Done()
			log.Printf("[Server] OnTimeoutRouteErrorHandler is called. err: %v", err)
		}),
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	start := time.Now()
	res, err := http.Get("http://localhost:8080/sleep?sleep=5")
	elapsed := time.Now().Sub(start)
	if err != nil {
		log.Printf("[Client] error occur. err: %v. elapsed: %v", err, elapsed)
		return
	}
	defer res.Body.Close()
	b, _ := ioutil.ReadAll(res.Body)
	log.Printf("[Client] StatusCode: %d, Body: %s, Elapsed: %v", res.StatusCode, string(b), elapsed)
	errHandleWait.Wait()
	// Output
	//2021/09/30 23:10:41 ## [Server:handleSleep] sleep 5 secs
	//2021/09/30 23:10:44 [Server] ctx.Request.Context is done
	//2021/09/30 23:10:44 [Client] error occur. err: Get "http://localhost:8080/sleep?sleep=5": EOF. elapsed: 3.004010498s
}

func newTimeoutHandler(onTimeoutErrHandler func(err error, c echo.Context)) *echo.Echo {
	e := echo.New()
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:                    nil,
		ErrorMessage:               "timeout",
		OnTimeoutRouteErrorHandler: onTimeoutErrHandler,
		Timeout:                    3 * time.Second,
	}), func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			go func() {
				select {
				case <-ctx.Request().Context().Done():
					log.Printf("[Server] ctx.Request.Context is done")
				}
			}()
			return next(ctx)
		}
	})
	e.GET("/sleep", func(ctx echo.Context) error {
		sleep := 0
		if sleepVal := ctx.QueryParam("sleep"); sleepVal != "" {
			s, err := strconv.Atoi(sleepVal)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, echo.Map{
					"message": fmt.Sprintf("Invalid query sleep=%s", sleepVal),
				})
			}
			sleep = s
		}

		log.Printf("## [Server:handleSleep] sleep %d secs", sleep)
		if sleep != 0 {
			time.Sleep(time.Duration(sleep) * time.Second)
		}
		return ctx.JSON(http.StatusOK, echo.Map{
			"sleep": sleep,
		})
	})
	return e
}
