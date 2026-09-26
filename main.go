package main

import (
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()


	server := &http.Server{
		Addr: ":8080",
		Handler: serverMux,
	}


	serverMux.Handle("/app/",http.StripPrefix("/app/",http.FileServer(http.Dir("."))))

	serverMux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request){
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})


	server.ListenAndServe()

}
