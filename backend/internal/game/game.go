package game

import (
	"encoding/json"
	"math/rand"
	"time"
)

type Player struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsBot    bool   `json:"isBot"`
}

type Game struct {
	ID            string    `json:"id"`
	Board         [6][7]int `json:"board"` // 7 columns, 6 rows
	Players       [2]Player `json:"players"`
	CurrentPlayer int       `json:"currentPlayer"` // 0 or 1
	Status        string    `json:"status"`        // "waiting", "playing", "finished"
	Winner        int       `json:"winner"`        // -1: draw, 0/1: player index
	CreatedAt     time.Time `json:"createdAt"`
	LastMoveAt    time.Time `json:"lastMoveAt"`
}

type Move struct {
	PlayerID string `json:"playerId"`
	Column   int    `json:"column"`
	Row      int    `json:"row"`
}

func NewGame(id string, player1 Player) *Game {
	return &Game{
		ID:        id,
		Board:     [6][7]int{},
		Players:   [2]Player{player1, {}},
		Status:    "waiting",
		Winner:    -1,
		CreatedAt: time.Now(),
	}
}

func (g *Game) AddPlayer(player2 Player) {
	g.Players[1] = player2
	g.Status = "playing"
	g.CurrentPlayer = rand.Intn(2) // Random starting player
}

func (g *Game) AddBot() {
	g.Players[1] = Player{
		ID:       "bot_" + g.ID,
		Username: "CompetitiveBot",
		IsBot:    true,
	}
	g.Status = "playing"
	g.CurrentPlayer = rand.Intn(2)
}

func (g *Game) MakeMove(column int) (bool, int, error) {
	if g.Status != "playing" {
		return false, -1, &GameError{"game is not active"}
	}

	if column < 0 || column >= 7 {
		return false, -1, &GameError{"invalid column"}
	}

	// Find the lowest available row in the column
	row := -1
	for r := 5; r >= 0; r-- {
		if g.Board[r][column] == 0 {
			row = r
			break
		}
	}

	if row == -1 {
		return false, -1, &GameError{"column is full"}
	}

	// Place the disc (1 for player1, 2 for player2)
	g.Board[row][column] = g.CurrentPlayer + 1
	g.LastMoveAt = time.Now()

	// Check for win
	if g.CheckWin(row, column) {
		g.Status = "finished"
		g.Winner = g.CurrentPlayer
		return true, row, nil
	}

	// Check for draw
	if g.IsBoardFull() {
		g.Status = "finished"
		g.Winner = -1 // Draw
		return true, row, nil
	}

	// Switch player
	g.CurrentPlayer = 1 - g.CurrentPlayer
	return true, row, nil
}

func (g *Game) CheckWin(row, col int) bool {
	player := g.Board[row][col]
	if player == 0 {
		return false
	}

	// Check horizontal
	count := 0
	for c := 0; c < 7; c++ {
		if g.Board[row][c] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check vertical
	count = 0
	for r := 0; r < 6; r++ {
		if g.Board[r][col] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	// Check diagonal (top-left to bottom-right)
	count = 0
	startRow, startCol := row, col
	for startRow > 0 && startCol > 0 {
		startRow--
		startCol--
	}

	for startRow < 6 && startCol < 7 {
		if g.Board[startRow][startCol] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
		startRow++
		startCol++
	}

	// Check diagonal (top-right to bottom-left)
	count = 0
	startRow, startCol = row, col
	for startRow > 0 && startCol < 6 {
		startRow--
		startCol++
	}

	for startRow < 6 && startCol >= 0 {
		if g.Board[startRow][startCol] == player {
			count++
			if count == 4 {
				return true
			}
		} else {
			count = 0
		}
		startRow++
		startCol--
	}

	return false
}

func (g *Game) IsBoardFull() bool {
	for c := 0; c < 7; c++ {
		if g.Board[0][c] == 0 {
			return false
		}
	}
	return true
}

func (g *Game) ToJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Game) GetCurrentPlayer() Player {
	return g.Players[g.CurrentPlayer]
}

// Simple error type for game package
type GameError struct {
	Message string
}

func (e *GameError) Error() string {
	return e.Message
}