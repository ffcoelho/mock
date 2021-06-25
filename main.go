package main

import (
	"fmt"
	"net/http"
)

const port int = 9000

func main() {
	http.HandleFunc("/", handler)

	fmt.Printf("Mock started. Listening on http://localhost:%d\n", port)

	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Hi")
}
