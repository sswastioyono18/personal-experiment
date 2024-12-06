package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello from Golang App!")
		fmt.Println("test logging, incoming request at ", time.Now())
	})
	fmt.Println("Golang app is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
