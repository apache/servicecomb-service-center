package main

import (
	"github.com/tedsuo/rata"
	"net/http"
)

func handler2(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(""))
}
func main() {
	petRoutes := rata.Routes{
		{Name: "get_pet", Method: "GET", Path: "/"},
	}
	petHandlers := rata.Handlers{
		"get_pet":    http.HandlerFunc(handler2),
	}
	router, err := rata.NewRouter(petRoutes, petHandlers)
	if err != nil {
		panic(err)
	}

	// The router is just an http.Handler, so it can be used to create a server in the usual fashion:
	http.Handle("/", router)
	http.ListenAndServe(":9980", nil)
}