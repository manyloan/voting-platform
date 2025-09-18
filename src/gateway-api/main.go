// src/gateway-api/main.go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// newReverseProxy cria um novo reverse proxy para um serviço de destino.
func newReverseProxy(target string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(url), nil
}

func main() {
	// Cria os proxies para cada um dos nossos serviços internos
	pollsProxy, err := newReverseProxy("http://polls-api:8080")
	if err != nil {
		log.Fatal(err)
	}

	voteProxy, err := newReverseProxy("http://vote-processor:8080")
	if err != nil {
		log.Fatal(err)
	}

	resultsProxy, err := newReverseProxy("http://results-api:8080")
	if err != nil {
		log.Fatal(err)
	}

	// O handler principal que vai rotear o tráfego
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Recebida requisição: %s %s", r.Method, r.URL.Path)

		// Roteia baseado no prefixo da URL
		if strings.HasPrefix(r.URL.Path, "/api/polls") {
			// Remove o prefixo /api para que o serviço de destino receba o caminho correto
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			pollsProxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/votes") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			voteProxy.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/results") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
			resultsProxy.ServeHTTP(w, r)
			return
		}

		// Responde com 404 Not Found se nenhuma rota corresponder
		http.NotFound(w, r)
	})

	// Mantém os endpoints de health e metrics
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Println("API Gateway iniciado na porta 8080, roteando tráfego...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
