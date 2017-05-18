package main

import (
"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}