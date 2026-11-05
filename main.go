package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	html := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
	w.Write([]byte(html))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{
		"error": msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)

}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits.Store(0)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	c := chirp{}
	err := decoder.Decode(&c)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	if utf8.RuneCountInString(c.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleaned_chirp := cleanBody(c.Body)

	respondWithJSON(w, http.StatusOK, map[string]string{
		"cleaned_body": cleaned_chirp,
	})
}

func cleanBody(body string) string {
	words := strings.Fields(body)
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	for i, word := range words {
		for _, bad := range badWords {
			if strings.ToLower(word) == bad {
				words[i] = "****"
				break
			}
		}
	}
	return strings.Join(words, " ")
}

func main() {
	const PORT string = ":8080"

	config := &apiConfig{}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/metrics", config.getMetrics)
	mux.HandleFunc("POST /admin/reset", config.resetMetrics)

	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	fileServer := http.FileServer(http.Dir("."))
	appHandler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/", config.middlewareMetricsInc(appHandler))

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Printf("Server listening on %s", PORT)
	log.Fatal(s.ListenAndServe())
}
