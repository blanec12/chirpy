package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{
		"error": msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)

}

func main() {
	config, err := NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/metrics", config.getMetrics)
	mux.HandleFunc("POST /admin/reset", config.reset)

	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/login", config.loginUser)
	mux.HandleFunc("POST /api/refresh", config.refreshAccessToken)
	mux.HandleFunc("POST /api/revoke", config.revokeRefreshToken)
	mux.HandleFunc("POST /api/users", config.createUser)
	mux.HandleFunc("PUT /api/users", config.updateUser)
	mux.HandleFunc("POST /api/chirps", config.createChirp)
	mux.HandleFunc("GET /api/chirps", config.getAllChirps)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", config.deleteChirp)
	mux.HandleFunc("GET /api/chirps/{chirpID}", config.getChirp)

	mux.HandleFunc("POST /api/polka/webhooks", config.processWebhook)

	fileServer := http.FileServer(http.Dir("."))
	appHandler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/", config.middlewareMetricsInc(appHandler))

	s := &http.Server{
		Addr:    config.port,
		Handler: mux,
	}

	fmt.Printf("Server listening on %s", config.port)
	log.Fatal(s.ListenAndServe())
}
