package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	FileserverHits atomic.Int32
}

func (c *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) GetMetrics() string {
	hits := c.FileserverHits.Load()
	return fmt.Sprintf("%v", hits)
}

func (c *apiConfig) ResetMetrics() bool {
	hits := c.FileserverHits.Load()
	success := c.FileserverHits.CompareAndSwap(hits, 0)
	return success
}

func updateHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}
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

	apiCfg := &apiConfig{}
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(updateHeader(http.StripPrefix("/app", http.FileServer(http.Dir("./app"))))))

	fs := http.FileServer(http.Dir("./app/assets/"))
	mux.Handle("/app/assets", http.StripPrefix("/app/assets", fs))

	// get metrics
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		hits := apiCfg.GetMetrics()
		_, err := w.Write([]byte("Hits: " + string(hits)))
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = apiCfg.ResetMetrics()
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	})

	srv := &http.Server{Addr: ":8080", Handler: &mux}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
