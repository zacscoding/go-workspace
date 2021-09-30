package timeout

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// Test_ServerDefaultTimeout_ClientTimeout tests with below configs.
// 1. Server: http.Server.WriteTimeout: 3s
// 2. Client: request timeout: 2s | server sleep: 5s
// ==> Server's gin.Request.Context will be done after 2s (i.e. closed client connection)
// ==> Client will be received context deadline error after 2s.
func Test_ServerDefaultTimeout_ClientTimeout(t *testing.T) {
	runTests(2*time.Second, 5)
	// Output
	//2021/09/30 20:28:33 ## [Client] Do request http://localhost:3000/sleep?sleep=5
	//2021/09/30 20:28:33 ## [Server:handleSleep] sleep 5 secs
	//2021/09/30 20:28:35 [Server:firstMiddleware] gin.Request.Context is done. elapsed: 2.000353908s
	//2021/09/30 20:28:35 [Test] Sleep: 5secs > code:0, header:map[], body:, err:Get "http://localhost:3000/sleep?sleep=5": context deadline exceeded (Client.Timeout exceeded while awaiting headers), elapsed: 2.003481356s
	//2021/09/30 20:28:38 [GIN] 2021/09/30 - 20:28:38 | 200 [closed: true] |  5.001099575s |             ::1 | GET     "/sleep?sleep=5"
}

// Test_ServerDefaultTimeout_ClientNoTimeout tests with below configs.
// [Server]: http.Server.WriteTimeout: 3s
// [Client]: request timeout: none | server sleep: 5s
// ==> Client will be received EOF error after 5s.
// ==> Server's gin.Request.Context will be done after 5s.
// i.e. Server cloud not respond 200 Status OK because WriteTimeout seted 3s and client will be responded after 5s.
func Test_ServerDefaultTimeout_ClientNoTimeout(t *testing.T) {
	runTests(0, 5)
	// Output
	//2021/09/30 20:39:45 ## [Client] Do request http://localhost:3000/sleep?sleep=5
	//2021/09/30 20:39:45 ## [Server:handleSleep] sleep 5 secs
	//2021/09/30 20:39:50 [GIN] 2021/09/30 - 20:39:50 | 200 [closed: false] |  5.001960573s |             ::1 | GET     "/sleep?sleep=5"
	//2021/09/30 20:39:50 [Server:firstMiddleware] gin.Request.Context is done. elapsed: 5.002088081s
	//2021/09/30 20:39:50 [Test] Sleep: 5secs > code:0, header:map[], body:, err:Get "http://localhost:3000/sleep?sleep=5": EOF, elapsed: 5.006929836s
}

// Test_ServerTimeoutMiddleware1_ClientNoTimeout tests with below configs.
// [Server]: http.Server.WriteTimeout: 3s, middleware timeout: 5s (greater than WriteTimeout)
// [Client]: request timeout: none | server sleep: 6s
// ==> Client will be received EOF error after 5s (because of server's timeout middleware)
// ==> Server's gin.Request.Context will be done after 5s.
// i.e. Server cloud not respond 200 Status OK because of WriteTimeout seted 3s and client will be responded after 5s not 6s.
func Test_ServerTimeoutMiddleware1_ClientNoTimeout(t *testing.T) {
	runTests(0, 6, timeoutMiddleware(5*time.Second))
	// Output
	//2021/09/30 20:46:00 ## [Client] Do request http://localhost:3000/sleep?sleep=6
	//2021/09/30 20:46:00 ## [Server:handleSleep] sleep 6 secs
	//2021/09/30 20:46:05 [Server:handleRequest] context in request is done
	//2021/09/30 20:46:05 [Server:timeoutMiddleware] context deadline exceeded
	//2021/09/30 20:46:05 [GIN] 2021/09/30 - 20:46:05 | 504 [closed: true] |  5.004188542s |             ::1 | GET     "/sleep?sleep=6"
	//2021/09/30 20:46:05 [Server:firstMiddleware] gin.Request.Context is done. elapsed: 5.004348554s
	//2021/09/30 20:46:05 [Test] Sleep: 6secs > code:0, header:map[], body:, err:Get "http://localhost:3000/sleep?sleep=6": EOF, elapsed: 5.009290826s
}

// Test_ServerTimeoutMiddleware1_ClientNoTimeout tests with below configs.
// [Server]: http.Server.WriteTimeout: 3s, middleware timeout: 2s (less than WriteTimeout)
// [Client]: request timeout: none | server sleep: 6s
// ==> Client will be received 504 timeout error after 2s (because of server's timeout middleware)
// ==> Server's gin.Request.Context will be done after 2s.
func Test_ServerTimeoutMiddleware2_ClientNoTimeout(t *testing.T) {
	runTests(0, 6, timeoutMiddleware(2*time.Second))
	// Output
	//2021/09/30 20:48:12 ## [Client] Do request http://localhost:3000/sleep?sleep=6
	//2021/09/30 20:48:12 ## [Server:handleSleep] sleep 6 secs
	//2021/09/30 20:48:14 [Server:handleRequest] context in request is done
	//2021/09/30 20:48:14 [Server:timeoutMiddleware] context deadline exceeded
	//2021/09/30 20:48:14 [GIN] 2021/09/30 - 20:48:14 | 504 [closed: true] |  2.005019247s |             ::1 | GET     "/sleep?sleep=6"
	//2021/09/30 20:48:14 [Server:firstMiddleware] gin.Request.Context is done. elapsed: 2.005157184s
	//2021/09/30 20:48:14 [Test] Sleep: 6secs > code:504, header:map[Content-Length:[18] Content-Type:[text/plain; charset=utf-8] Date:[Thu, 30 Sep 2021 11:48:14 GMT] First:[second]], body:{"err":"timeout!"}, err:<nil>, elapsed: 2.009638673s
}

func runTests(clientRequestTimeout time.Duration, serverSleepSecs int, middlewares ...gin.HandlerFunc) {
	go StartServer(middlewares...)
	time.Sleep(time.Second)

	start := time.Now()
	code, header, body, err := requestWithSleepSecs(clientRequestTimeout, serverSleepSecs)
	elapsed := time.Now().Sub(start)
	log.Printf("[Test] Sleep: %dsecs > code:%d, header:%v, body:%s, err:%v, elapsed: %v", serverSleepSecs, code, header, body, err, elapsed)
	time.Sleep(10 * time.Second)
}

func requestWithSleepSecs(requestTimeout time.Duration, serverSleepSecs int) (int, http.Header, string, error) {
	url := "http://localhost:3000/sleep"
	if serverSleepSecs > 0 {
		url += "?sleep=" + strconv.Itoa(serverSleepSecs)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, "", err
	}
	client := http.Client{}
	if requestTimeout != 0 {
		client.Timeout = requestTimeout
	}

	log.Printf("## [Client] Do request %s", url)
	resp, err := client.Do(request)
	if err != nil {
		var (
			header http.Header
			code   = 0
		)
		if resp != nil {
			header = resp.Header
			code = resp.StatusCode
		}
		return code, header, "", err
	}

	var (
		header http.Header
		code   = 0
	)
	if resp != nil {
		defer resp.Body.Close()
		header = resp.Header
		code = resp.StatusCode
	}
	b, _ := ioutil.ReadAll(resp.Body)
	return code, header, string(b), nil
}
