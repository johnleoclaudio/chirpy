package main

import (
	"net/http"
)

func main() {
	mux := http.ServeMux{}
	srv := &http.Server{Addr: ":8080", Handler: &mux}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
