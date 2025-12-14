package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.ServeMux{}

	// readiness endpoint
	// returns 200 OK if the server is ready to accept requests
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.Handle("/app", http.StripPrefix("/app", http.FileServer(http.Dir("./app"))))

	fs := http.FileServer(http.Dir("./app/assets/"))
	mux.Handle("/app/assets", http.StripPrefix("/app/assets", fs))

	srv := &http.Server{Addr: ":8080", Handler: &mux}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
