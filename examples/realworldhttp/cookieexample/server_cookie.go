package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// curl -XGET -c cookie.txt -b cookie.txt -b "key=value" http://localhost:8080/cookie
	http.HandleFunc("/cookie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Set-Cookie", "VISIT=TRUE")
		if values, ok := r.Header["Cookie"]; ok {
			log.Printf("Exist Cookie: %v", values)
			fmt.Fprintf(w, "<html><body>Second!</body></html>\n")
		} else {
			log.Println("Empty Cookie")
			fmt.Fprintf(w, "<html><body>First!</body></html>\n")
		}
	})
	http.ListenAndServe(":8080", nil)
}
