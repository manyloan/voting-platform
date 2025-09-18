// src/gateway-api/main.go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

func newReverseProxy(target string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(url), nil
}

func main() {
	pollsProxy, _ := newReverseProxy("http://polls-api:8080")
	voteProxy, _ := newReverseProxy("http://vote-processor:8080")
	resultsProxy, _ := newReverseProxy("http://results-api:8080")

	router := http.NewServeMux()

	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recebida requisição: %s %s", r.Method, r.URL.Path)

		if strings.HasPrefix(r.URL.Path, "/api/polls") || strings.HasPrefix(r.URL.Path, "/polls") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			pollsProxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/votes") || strings.HasPrefix(r.URL.Path, "/votes") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			voteProxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/results-hub") || strings.HasPrefix(r.URL.Path, "/api/results") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			resultsProxy.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:30000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := corsHandler.Handler(router)

	log.Println("API Gateway iniciado na porta 8080, com CORS ativado para localhost:30000...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
