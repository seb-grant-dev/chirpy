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
	platform string
	jwtSecret string
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
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("403 Forbidden\n"))
		return
	}
	cfg.fileserverHits.Store(0)
	cfg.DB.ResetUsers(req.Context())
	respondWithJSON(w, 200, "Stats and users reset.")
}

func main() {

	godotenv.Load()
	serverMux := http.NewServeMux()

	apiCfg := &apiConfig{}
	userHandler := &UserHandler{}
	chirpHandler := &ChirpHandler{}

	server := &http.Server{
		Addr: ":8080",
		Handler: serverMux,
	}

	apiCfg.platform = os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres",dbUrl)
	if err != nil {
		fmt.Errorf("FATAL: Could not connect to database [%s]\n",err)
		os.Exit(1)
	}
	
	dbQueries := database.New(db)
	apiCfg.DB = dbQueries
	apiCfg.jwtSecret = jwtSecret
	userHandler.DB = dbQueries
	userHandler.jwtSecret = jwtSecret
	chirpHandler.DB = dbQueries
	chirpHandler.jwtSecret = jwtSecret

	webhookHandler := &WebhookHandler{
		DB: dbQueries,
		jwtSecret: jwtSecret,
	}

	handler := http.StripPrefix("/app/",http.FileServer(http.Dir(".")))
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request){
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	serverMux.HandleFunc("POST /api/users",userHandler.createUser)
	serverMux.HandleFunc("PUT /api/users", userHandler.updateUser)
	serverMux.HandleFunc("POST /api/login", userHandler.doLogin)
	serverMux.HandleFunc("POST /api/refresh", userHandler.refreshToken)
	serverMux.HandleFunc("POST /api/revoke", userHandler.revokeToken)

	serverMux.HandleFunc("GET /api/chirps",chirpHandler.getChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}",chirpHandler.getChirp)
	serverMux.HandleFunc("POST /api/chirps",chirpHandler.createChirp)
	serverMux.HandleFunc("DELETE /api/chirps/{chirpID}",chirpHandler.deleteChirp)
	

	serverMux.HandleFunc("POST /admin/reset",apiCfg.resetHits)
	serverMux.HandleFunc("GET /admin/metrics",apiCfg.outputHits)
	
	serverMux.HandleFunc("POST /api/polka/webhooks",webhookHandler.handle)

	server.ListenAndServe()

}
