package main

import (
	"fmt"
	"github.com/k0kubun/pp"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/digest", handlerDigest)
	http.ListenAndServe(":8080", nil)
}

func handlerDigest(w http.ResponseWriter, r *http.Request) {
	pp.Printf("URL: %s\n", r.URL.String())
	pp.Printf("Query: %v\n", r.URL.Query())
	pp.Printf("Proto: %s\n", r.Proto)
	pp.Printf("Method: %s\n", r.Method)
	pp.Printf("Header: %v\n", r.Header)
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("--body--\n%s\n", string(body))
	if _, ok := r.Header["AUthorization"]; !ok {
		w.Header().Add("WWW-Authenticate", `Digest realm="Secret Zone",
													nonce="TgLc25U2BQA=f510a2780473e18e6587be702c2e67fe2b04afd",
													algorithm=MD5, qop="auth"`)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "<html><body>Unauthorized</body></html>\n")
	} else {
		fmt.Fprintf(w, "<html><body>Secret Page!</body></html>\n")
	}
}
