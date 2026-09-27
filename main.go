package main

import (
	"fmt"
	"sync/atomic"
	"net/http"
)
type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	fmt.Println("Applying tracking middleware")
	return http.HandlerFunc(func (w http.ResponseWriter, req *http.Request) {
			fmt.Println("App Hit. Incrementing")
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w,req)
		})
}

func (cfg *apiConfig) outputHits(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(fmt.Sprintf("Hits: %d",cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetHits(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits.Store(0)
}

func main() {
	serverMux := http.NewServeMux()

	apiCfg := &apiConfig{}

	server := &http.Server{
		Addr: ":8080",
		Handler: serverMux,
	}

	handler := http.StripPrefix("/app/",http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	serverMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request){
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	serverMux.HandleFunc("/metrics",apiCfg.outputHits)
	serverMux.HandleFunc("/reset",apiCfg.resetHits)

	server.ListenAndServe()

}
