package main

import (
	"connect-four/internal/database"
	"connect-four/internal/game"
	"connect-four/internal/websockethub"  // Use the renamed package
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type Server struct {
	hub   *websockethub.Hub  // Use websockethub, not websocket
	store *database.PostgresStore
}

func NewServer() *Server {
	// Initialize database
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://user:password@localhost:5432/connectfour?sslmode=disable"
	}

	store, err := database.NewPostgresStore(connStr)
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v", err)
		store = nil
	} else {
		if err := store.Init(); err != nil {
			log.Printf("Warning: Could not initialize database: %v", err)
		}
	}

	return &Server{
		hub:   websockethub.NewHub(),  // Use websockethub
		store: store,
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &websockethub.Client{  // Use websockethub
		Hub:      s.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		PlayerID: r.URL.Query().Get("username") + "_" + generatePlayerID(),
		GameID:   r.URL.Query().Get("gameId"),
	}

	s.hub.Register <- client

	// Start goroutines for this client
	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) handleCreateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	player := game.Player{
		ID:       generatePlayerID(),
		Username: req.Username,
		IsBot:    false,
	}

	game := s.hub.CreateGame(player)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func (s *Server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.store == nil {
		// Return empty array if database is not available
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	entries, err := s.store.GetLeaderboard()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entries)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	rand.Seed(time.Now().UnixNano())
	server := NewServer()

	// Start WebSocket hub
	go server.hub.Run()

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	})

	// Routes
	http.Handle("/health", c.Handler(http.HandlerFunc(server.handleHealth)))
	http.Handle("/ws", c.Handler(http.HandlerFunc(server.handleWebSocket)))
	http.Handle("/game/create", c.Handler(http.HandlerFunc(server.handleCreateGame)))
	http.Handle("/leaderboard", c.Handler(http.HandlerFunc(server.handleLeaderboard)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func generatePlayerID() string {
	return string(rune(rand.Intn(10000)))
}