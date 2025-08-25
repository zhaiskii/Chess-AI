package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	chessService := NewChessService()
	aiService := NewAIService()
	handlers := NewHandlers(chessService, aiService)
	log.Println("Hello2");

	r := mux.NewRouter()

	r.Use(corsMiddleware)
	r.Use(loggingMiddleware)

	r.HandleFunc("/health", handlers.Health).Methods("GET")

	api := r.PathPrefix("/api").Subrouter()
	
	api.Use(corsMiddleware)

	api.HandleFunc("/game", handlers.GetGameState).Methods("GET")
	api.HandleFunc("/move", handlers.MakeMove).Methods("POST", "OPTIONS")
	api.HandleFunc("/new-game", handlers.NewGame).Methods("POST")
	api.HandleFunc("/valid-moves", handlers.GetValidMoves).Methods("GET")
	api.HandleFunc("/change-depth", handlers.ChangeDepth).Methods("POST", "OPTIONS")

	api.HandleFunc("/ai/move", handlers.ForceAIMove).Methods("POST")
	api.HandleFunc("/ai/stats", handlers.GetAIStats).Methods("GET")
	api.HandleFunc("/ai/difficulty", handlers.SetDifficulty).Methods("POST")
	
	api.HandleFunc("/evaluate", handlers.EvaluatePosition).Methods("GET")
	api.HandleFunc("/history", handlers.GetGameHistory).Methods("GET")

	port := getEnv("PORT", "8080")
	
	log.Printf("Chess AI server starting on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("   GET  /health")
	log.Printf("   GET  /api/game")
	log.Printf("   POST /api/move")
	log.Printf("   POST /api/new-game")
	log.Printf("   POST /api/ai/move")
	log.Printf("   POST /api/change-depth");
	
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.WriteHeader(http.StatusOK)
			return
		}

		//fix that later

		log.Printf("request %s", r.URL.Path);

		origin := r.Header.Get("Origin")
		if origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapped, r)
		
		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.RequestURI, wrapped.statusCode, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}