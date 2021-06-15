package main

import (
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	var (
		port     = 8900
		endpoint = fmt.Sprintf("http://localhost:%d", port)
	)

	go func() {
		if err := StartServer(fmt.Sprintf(":%d", port)); err != nil {
			log.Fatalln(err)
		}
	}()
	// (1) Call /public no require auth
	u := fmt.Sprintf("%s/public", endpoint)
	resp, err := http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	printResponse("", http.MethodGet, u, resp)

	// (2) Call /auth without auth token
	u = fmt.Sprintf("%s/auth", endpoint)
	resp, err = http.Get(u)
	if err != nil {
		log.Fatal(err)
	}
	printResponse("without auth token", http.MethodGet, u, resp)

	// (3) Signin with invalid user
	u = fmt.Sprintf("%s/signin", endpoint)
	resp, err = http.PostForm(u, url.Values{
		"username": []string{"zac"},
		"password": []string{"invalid"},
	})
	if err != nil {
		log.Fatal(err)
	}
	printResponse("invalid user info", http.MethodPost, u, resp)

	// (4) Signin with valid user
	u = fmt.Sprintf("%s/signin", endpoint)
	resp, err = http.PostForm(u, url.Values{
		"username": []string{"zac"},
		"password": []string{"coding"},
	})
	if err != nil {
		log.Fatal(err)
	}
	body := printResponse("signin with valid user", http.MethodPost, u, resp)
	parsed, err := gabs.ParseJSON(body)
	if err != nil {
		log.Fatal(err)
	}
	token := parsed.Path("token").Data().(string)

	// (5) Call /auth with auth token
	u = fmt.Sprintf("%s/auth", endpoint)
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	printResponse("", http.MethodGet, u, resp)

	// Output
	//2021/06/16 00:17:22 [Client - ] Call GET http://localhost:8900/public > code: 200, body:{"message":"Hello :)"}
	//2021/06/16 00:17:22 [Echo] BeforeFunc is called
	//2021/06/16 00:17:22 [Echo] ErrorHandler is called. err: code=400, message=missing or malformed jwt
	//2021/06/16 00:17:22 [Client - without auth token] Call GET http://localhost:8900/auth > code: 401, body:{"message":"Unauthorized"}
	//2021/06/16 00:17:22 [Client - invalid user info] Call POST http://localhost:8900/signin > code: 401, body:{"message":"Unauthorized"}
	//2021/06/16 00:17:22 [Client - signin with valid user] Call POST http://localhost:8900/signin > code: 200, body:{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiemFjIiwiYWRtaW4iOnRydWUsImV4cCI6MTYyNDAyOTQ0Mn0.yoAp6KF7bbT62dsaFiY9M7DSXI22VXlTRz1xUiWODuQ"}
	//2021/06/16 00:17:22 [Echo] BeforeFunc is called
	//2021/06/16 00:17:22 [Echo] SuccessHandler is called. name: zac
	//2021/06/16 00:17:22 [Client - ] Call GET http://localhost:8900/auth > code: 200, body:{"isAdmin":true,"name":"zac"}
}

func printResponse(message, method, endpoint string, resp *http.Response) []byte {
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[Client - %s] Call %s %s > code: %d, body:%s", message, strings.ToUpper(method), endpoint, resp.StatusCode, string(body))
	return body
}
