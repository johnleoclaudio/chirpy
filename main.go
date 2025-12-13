package main

import (
	"net/http"
)

func main() {
	mux := http.ServeMux{}

	mux.Handle("/", http.FileServer(http.Dir(".")))

	srv := &http.Server{Addr: ":8080", Handler: &mux}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
