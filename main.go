package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"os"
	"sync/atomic"
	"net/http"
	"database/sql"
	"github.com/joho/godotenv"
	"github.com/seb-grant-dev/chirpy/internal/database"
)
type apiConfig struct {
	fileserverHits atomic.Int32
	DB *database.Queries
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

	godotenv.Load()
	serverMux := http.NewServeMux()

	apiCfg := &apiConfig{}

	server := &http.Server{
		Addr: ":8080",
		Handler: serverMux,
	}

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres",dbUrl)
	if err != nil {
		dbQueries := database.New(db)
		apiCfg.DB = dbQueries
	}

	handler := http.StripPrefix("/app/",http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request){
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	serverMux.HandleFunc("POST /api/validate_chirp",validateChirp)
	
	serverMux.HandleFunc("POST /admin/reset",apiCfg.resetHits)
	serverMux.HandleFunc("GET /admin/metrics",apiCfg.outputHits)
	









	server.ListenAndServe()

}
