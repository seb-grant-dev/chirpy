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
	metricsTemplate := `
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`
	w.Write([]byte(fmt.Sprintf(metricsTemplate,cfg.fileserverHits.Load())))
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

	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request){
		w.WriteHeader(200)
		w.Write([]byte("OK\n"))
	})

	serverMux.HandleFunc("POST /admin/reset",apiCfg.resetHits)
	serverMux.HandleFunc("GET /admin/metrics",apiCfg.outputHits)
	server.ListenAndServe()

}
