package main

import (
	"fmt"
	"log"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello!")
}

func worldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "World!")
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Test!")
}

func main() {
	http.Handle("/hello", http.HandlerFunc(helloHandler))
	http.Handle("/world", http.HandlerFunc(worldHandler))
	http.Handle("/v1/test", http.HandlerFunc(testHandler))

	log.Println("Open URL http://0.0.0.0:8081/ in your browser.")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}
