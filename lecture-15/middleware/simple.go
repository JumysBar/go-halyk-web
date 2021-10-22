package main

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello world!")
	return
}

func SimpleMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Simple middleware message")
		next(w, r)
	}
}

func main() {
	http.HandleFunc("/", SimpleMiddleware(Handler))
	http.ListenAndServe(":8080", nil)
}
