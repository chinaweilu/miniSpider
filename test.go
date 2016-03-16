package main

import (
	"fmt"
	"net/http"
	"runtime/pprof"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	p := pprof.Lookup("goroutine")
	p.WriteTo(w, 1)
}

func main() {
	ch := make(chan int, 10)
	for i := 0; i < 10000; i++ {
		go Task(i, ch)
		<-ch
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8989", nil)
}

func Task(i int, ch chan int) {
	fmt.Println(i)
	ch <- 1
}
