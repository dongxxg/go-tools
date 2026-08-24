package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func fetch(url string, ch chan<- string) {
	start := time.Now()
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		ch <- fmt.Sprint(err)
		return
	}
	nbytes, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		ch <- fmt.Sprintf("While reading %s: %v", url, err)
		return
	}
	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("%.2fs %7d %s", secs, nbytes, url)
}

// HttpRequestUrls concurrent request
func HttpRequestUrls() {
	start := time.Now()
	ch := make(chan string)
	urls := [...]string{"https://www.baidu.com", "https://shanexiang.github.com"}
	for _, url := range urls {
		go fetch(url, ch)
	}
	for range urls {
		fmt.Println(<-ch)
	}
	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

var (
	mu      sync.Mutex
	counter int
)

func handler(w http.ResponseWriter, r *http.Request) {
	defer mu.Unlock()
	mu.Lock()
	if r.URL.RequestURI() == "/favicon.ico" {
		return
	}
	counter++
	log.Printf("Path %s was visited %d", r.URL.Path, counter)
	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
}

func counterHandler(w http.ResponseWriter, r *http.Request) {
	defer mu.Unlock()
	mu.Lock()
	fmt.Fprintf(w, "URL.Path / was accessed %d\n", counter)
}

func httpHandler() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/counter", counterHandler)
	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

func main() {
	httpHandler()
}
