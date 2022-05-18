package main

import (
	"log"
	"net/http"
)

type server int

func (h *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	w.Write([]byte("hello, world!"))
}

func main() {
	var s server
	http.ListenAndServe("localhost:9999", &s)		//第一个参数为服务启动的地址，第二个参数为Handler, 任何实现了ServeHTTP方法的对象都可以作为HTTP的Handler
}
