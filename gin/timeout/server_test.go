package timeout

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestDefaultTimeout(t *testing.T) {
	go StartServer()
	time.Sleep(time.Second)

	sleep := 10
	code, header, body, err := requestSleep1(sleep)
	log.Printf("[Sleep: %dsecs] code:%d, header:%v, body:%s, err:%v\n", sleep, code, header, body, err)
	// Output
	//2020/09/22 14:08:41 ## [Client] Do request http://localhost:3000/sleep1?sleep=10
	//2020/09/22 14:08:41 ## [Server] sleep 10 secs
	//2020/09/22 14:08:51 [Sleep: 10secs] code:0, header:map[], body:, err:Get "http://localhost:3000/sleep1?sleep=10": EOF
}

// Server WriteTimeout :: 3 secs & Middleware timeout :: 5 secs
func TestTimeoutWithMiddlewareLessThanServerWriteTimeout(t *testing.T) {
	go StartServer(timeoutMiddleware(5 * time.Second))
	time.Sleep(time.Second)
	sleep := 10
	code, header, body, err := requestSleep1(sleep)
	log.Printf("[Sleep: %dsecs] code:%d, header:%v, body:%s, err:%v\n", sleep, code, header, body, err)
	// Output
	//2020/09/22 14:13:04 ## [Client] Do request http://localhost:3000/sleep1?sleep=10
	//2020/09/22 14:13:04 ## [Server] sleep 10 secs
	//2020/09/22 14:13:09 timeout occur in handleRequest()
	//err in timeout middleware defer: context deadline exceeded
	//2020/09/22 14:13:09 [Sleep: 10secs] code:0, header:map[], body:, err:Get "http://localhost:3000/sleep1?sleep=10": EOF
}

// Server WriteTimeout :: 3 secs & Middleware timeout :: 2 secs
func TestTimeoutWithMiddlewareGreaterThanServerWriteTimeout(t *testing.T) {
	go StartServer(timeoutMiddleware(2 * time.Second))
	time.Sleep(time.Second)
	sleep := 10
	code, header, body, err := requestSleep1(sleep)
	log.Printf("[Sleep: %dsecs] code:%d, header:%v, body:%s, err:%v\n", sleep, code, header, body, err)
	// Output
	//2020/09/22 14:30:51 ## [Client] Do request http://localhost:3000/sleep1?sleep=10
	//2020/09/22 14:30:51 ## [Server] sleep 10 secs
	//2020/09/22 14:30:53 timeout occur in handleRequest()
	//2020/09/22 14:30:53 [Sleep: 10secs] code:504, header:map[Content-Length:[18] Content-Type:[text/plain; charset=utf-8] Date:[Tue, 22 Sep 2020 05:30:53 GMT] First:[second]], body:{"err":"timeout!"}, err:<nil>
}

func requestSleep1(sleep int) (int, http.Header, string, error) {
	url := "http://localhost:3000/sleep1"
	if sleep > 0 {
		url += "?sleep=" + strconv.Itoa(sleep)
	}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, "", err
	}
	client := http.Client{}

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
