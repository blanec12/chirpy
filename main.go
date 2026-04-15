package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

type errorResponse struct {
	Error string `json:"error"`
}

type cleanedResponse struct {
	CleanedBody string `json:"cleaned_body"`
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("200 OK: server ready and listening"))
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "something went wrong",
		})
		return
	}

	if len(params.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{
			Error: "chirp is too long",
		})
		return
	}

	cleanedBody := cleanBody(params.Body)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cleanedResponse{
		CleanedBody: cleanedBody,
	})
}

func cleanBody(body string) string {
	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Fields(body)

	for i, word := range words {
		lowerWord := strings.ToLower(word)

		if badWords[lowerWord] {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	hits := cfg.fileServerHits.Load()

	html := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin!</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`, hits)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func main() {
	cfg := apiConfig{}
	mux := http.NewServeMux()

	fileserver := http.FileServer(http.Dir("."))

	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", fileserver)))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("failed to start http server: %v", err)
	}
}
