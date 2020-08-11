package main

import (
	"go-workspace/client/fastclient/fastarticle"
	"go-workspace/serverutil"
	"log"
)

// Scenario
// 1) Start server
// 2) Request to server success 	>> Circuit breaker {"Requests":1,"TotalSuccesses":1,"TotalFailures":0,"ConsecutiveSuccesses":0,"ConsecutiveFailures":0}
// 3) Request to server with fail   >> Circuit breaker {"Requests":2,"TotalSuccesses":1,"TotalFailures":1,"ConsecutiveSuccesses":0,"ConsecutiveFailures":1}
// >>> circuit closed -> open
// 4) Request to server : fast fail because of circuit breaker
// 5) Request to server : fast fail because of circuit breaker
//
// Output
// 2020/08/11 23:45:09 [#1] Try to get articles
// 2020/08/11 23:45:09 >> success
// 2020/08/11 23:45:09 [#2] Try to get articles
// [GIN] 2020/08/11 - 23:45:09 | 200 |            0s |       127.0.0.1 | GET      "/articles"
// 2020/08/11 23:45:10 ReadyToTrip: {"Requests":2,"TotalSuccesses":1,"TotalFailures":1,"ConsecutiveSuccesses":0,"ConsecutiveFailures":1}
// 2020/08/11 23:45:10 Circuit[article] closed -> open
// 2020/08/11 23:45:10 >> error : dial tcp4 127.0.0.1:4000: connectex: No connection could be made because the target machine actively refused it.
// 2020/08/11 23:45:10 [#3] Try to get articles
// 2020/08/11 23:45:10 >> error : circuit breaker is open
// 2020/08/11 23:45:10 [#3] Try to get articles
// 2020/08/11 23:45:10 >> error : circuit breaker is open
func main() {
	// start mock server
	s := serverutil.NewGinArticleServer()
	go func() {
		if err := s.Run(":3000"); err != nil {
			panic(err)
		}
	}()

	// client
	cli := fastarticle.NewArticleClient("http://localhost:3000")

	// success
	log.Println("[#1] Try to get articles")
	_, err := cli.GetArticles("", 0, 0)
	if err != nil {
		log.Println(">> error :", err.Error())
	} else {
		log.Println(">> success")
	}

	// fail
	log.Println("[#2] Try to get articles")
	_, err = cli.GetArticles("http://localhost:4000", 0, 0)
	if err != nil {
		log.Println(">> error :", err.Error())
	} else {
		log.Println(">> success")
	}

	// >>>>> Circuit open <<<<<

	// fail
	log.Println("[#3] Try to get articles")
	_, err = cli.GetArticles("http://localhost:3000", 0, 0)
	if err != nil {
		log.Println(">> error :", err.Error())
	} else {
		log.Println(">> success")
	}

	// fail
	log.Println("[#3] Try to get articles")
	_, err = cli.GetArticles("http://localhost:3000", 0, 0)
	if err != nil {
		log.Println(">> error :", err.Error())
	} else {
		log.Println(">> success")
	}
}
