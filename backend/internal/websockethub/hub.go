package websockethub

import (
	"connect-four/internal/game"
	"connect-four/internal/bot"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	PlayerID string
	GameID   string
}

type Hub struct {
	Clients    map[*Client]bool
	Games      map[string]*game.Game
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan Message
	Mutex      sync.RWMutex
}

type Message struct {
	Type    string          `json:"type"`
	Content json.RawMessage `json:"content"`
}

type GameMessage struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
	Column   int    `json:"column,omitempty"`
	Username string `json:"username,omitempty"`
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Games:      make(map[string]*game.Game),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	// Cleanup goroutine for abandoned games
	go h.cleanupRoutine()

	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.Mutex.Unlock()

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.Mutex.Unlock()

		case message := <-h.Broadcast:
			h.Mutex.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- h.formatMessage(message):
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.Mutex.RUnlock()
		}
	}
}

func (h *Hub) formatMessage(msg Message) []byte {
	jsonMsg, _ := json.Marshal(msg)
	return jsonMsg
}

func (h *Hub) CreateGame(player1 game.Player) *game.Game {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	gameID := generateGameID()
	newGame := game.NewGame(gameID, player1)
	h.Games[gameID] = newGame

	log.Printf("New game created: %s, Status: %s", gameID, newGame.Status)
	log.Printf("Player 1: %s (IsBot: %t)", player1.Username, player1.IsBot)

	// Start bot timeout
	go h.startBotTimeout(gameID)

	return newGame
}

func (h *Hub) JoinGame(gameID string, player2 game.Player) (*game.Game, error) {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	game, exists := h.Games[gameID]
	if !exists {
		return nil, &GameError{"game not found"}
	}

	if game.Status != "waiting" {
		return nil, &GameError{"game already started"}
	}

	game.AddPlayer(player2)
	h.broadcastGameUpdate(game)
	
	// If bot goes first after joining, trigger bot move
	if game.Status == "playing" && game.GetCurrentPlayer().IsBot {
		log.Printf("Bot goes first after joining, triggering bot move...")
		go h.makeBotMove(gameID)
	}
	
	return game, nil
}

func (h *Hub) MakeMove(gameID string, playerID string, column int) (*game.Game, error) {
	h.Mutex.Lock()

	game, exists := h.Games[gameID]
	if !exists {
		h.Mutex.Unlock()
		return nil, &GameError{"game not found"}
	}

	log.Printf("MakeMove called - Game: %s, Player: %s, Column: %d", gameID, playerID, column)

	// Check if it's player's turn
	currentPlayer := game.GetCurrentPlayer()
	if currentPlayer.ID != playerID {
		h.Mutex.Unlock()
		log.Printf("Not player's turn. Current player: %s, Requested player: %s", currentPlayer.ID, playerID)
		return nil, &GameError{"not your turn"}
	}

	success, _, err := game.MakeMove(column)
	if err != nil {
		h.Mutex.Unlock()
		log.Printf("Move error: %v", err)
		return nil, err
	}

	log.Printf("Move successful: %t, Game status: %s", success, game.Status)

	if success {
		// Broadcast the move result
		h.broadcastGameUpdate(game)
		
		// If bot's turn next, make bot move
		if game.Status == "playing" {
			nextPlayer := game.GetCurrentPlayer()
			log.Printf("Next player: %s (IsBot: %t)", nextPlayer.Username, nextPlayer.IsBot)
			
			if nextPlayer.IsBot {
				log.Printf("Triggering bot move after player move...")
				// Unlock mutex before starting bot goroutine
				h.Mutex.Unlock()
				go h.makeBotMove(gameID)
				return game, nil // Return here since we unlocked
			}
		}
	}

	h.Mutex.Unlock()
	return game, nil
}

func (h *Hub) startBotTimeout(gameID string) {
	time.Sleep(10 * time.Second)

	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	game, exists := h.Games[gameID]
	if !exists {
		return
	}

	// Check if still waiting for player
	if game.Status == "waiting" {
		log.Printf("Bot timeout reached for game %s, adding bot...", gameID)
		game.AddBot()
		h.broadcastGameUpdate(game)
		
		// If bot goes first, trigger bot move immediately
		if game.Status == "playing" && game.GetCurrentPlayer().IsBot {
			log.Printf("Bot goes first, triggering bot move...")
			go h.makeBotMove(gameID)
		}
	}
}

func (h *Hub) makeBotMove(gameID string) {
	time.Sleep(1 * time.Second) // Small delay for realism

	h.Mutex.Lock()
	defer h.Mutex.Unlock()

	log.Printf("Bot move starting for game: %s", gameID)

	game, exists := h.Games[gameID]
	if !exists {
		log.Printf("Game not found for bot move: %s", gameID)
		return
	}

	if game.Status != "playing" {
		log.Printf("Game not in playing state: %s", game.Status)
		return
	}

	currentPlayer := game.GetCurrentPlayer()
	if !currentPlayer.IsBot {
		log.Printf("Not bot's turn. Current player: %s", currentPlayer.Username)
		return
	}

	log.Printf("Bot calculating move for game %s...", gameID)
	bot := bot.NewBot()
	column := bot.CalculateMove(game)

	if column != -1 {
		log.Printf("Bot making move in column: %d", column)
		_, _, err := game.MakeMove(column)
		if err != nil {
			log.Printf("Bot move error: %v", err)
			return
		}
		h.broadcastGameUpdate(game)
		log.Printf("Bot move completed successfully")
		
		// If it's still bot's turn after move (shouldn't happen in normal game)
		if game.Status == "playing" && game.GetCurrentPlayer().IsBot {
			log.Printf("Bot gets another turn? Triggering again...")
			go h.makeBotMove(gameID)
		}
	} else {
		log.Printf("Bot couldn't find a valid move")
		// If bot can't find a move but it's still their turn, try center column as fallback
		log.Printf("Trying fallback to center column...")
		for col := 3; col >= 0 && col < 7; col = 3 + (3-col)*-1 { // Try center, then alternate sides
			if col != 3 {
				col = 3 - (col-3) // Alternate: 3, 2, 4, 1, 5, 0, 6
			}
			if h.isValidMove(game, col) {
				log.Printf("Fallback: Bot making move in column: %d", col)
				_, _, err := game.MakeMove(col)
				if err != nil {
					log.Printf("Fallback bot move error: %v", err)
				} else {
					h.broadcastGameUpdate(game)
					log.Printf("Fallback bot move completed successfully")
				}
				return
			}
		}
		log.Printf("No valid fallback moves found!")
	}
}

func (h *Hub) isValidMove(g *game.Game, column int) bool {
	if column < 0 || column >= 7 {
		return false
	}
	return g.Board[0][column] == 0
}

func (h *Hub) broadcastGameUpdate(g *game.Game) {
	gameJSON, err := json.Marshal(g)
	if err != nil {
		log.Printf("Error marshaling game update: %v", err)
		return
	}
	
	message := Message{
		Type:    "game_update",
		Content: gameJSON,
	}

	log.Printf("Broadcasting game update for game: %s", g.ID)
	h.Broadcast <- message
}

func (h *Hub) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.Mutex.Lock()
		for gameID, game := range h.Games {
			// Remove games older than 1 hour
			if time.Since(game.CreatedAt) > time.Hour {
				log.Printf("Cleaning up old game: %s", gameID)
				delete(h.Games, gameID)
			}
		}
		h.Mutex.Unlock()
	}
}

func generateGameID() string {
	return "game_" + time.Now().Format("20060102150405")
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		switch msg.Type {
		case "make_move":
			var moveMsg GameMessage
			if err := json.Unmarshal(msg.Content, &moveMsg); err != nil {
				log.Printf("Error unmarshaling move message: %v", err)
				continue
			}
			
			log.Printf("Received move message: GameID=%s, PlayerID=%s, Column=%d", 
				moveMsg.GameID, moveMsg.PlayerID, moveMsg.Column)
				
			game, err := c.Hub.MakeMove(moveMsg.GameID, moveMsg.PlayerID, moveMsg.Column)
			if err != nil {
				log.Printf("Move error: %v", err)
				// Send error back to client
				errorMsg := Message{
					Type: "error",
					Content: json.RawMessage(`{"message":"` + err.Error() + `"}`),
				}
				c.Send <- c.Hub.formatMessage(errorMsg)
			} else {
				log.Printf("Move processed successfully for game: %s", game.ID)
			}
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// Simple error type for hub
type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}