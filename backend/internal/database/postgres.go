package database

import (
	"connect-four/internal/game"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Init() error {
	query := `
	CREATE TABLE IF NOT EXISTS games (
		id VARCHAR(50) PRIMARY KEY,
		player1 VARCHAR(100) NOT NULL,
		player2 VARCHAR(100),
		winner VARCHAR(100),
		status VARCHAR(20) NOT NULL,
		board_state TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		finished_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS leaderboard (
		username VARCHAR(100) PRIMARY KEY,
		wins INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		draws INTEGER DEFAULT 0,
		updated_at TIMESTAMP NOT NULL
	);
	`

	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) SaveGame(g *game.Game) error {
	query := `
	INSERT INTO games (id, player1, player2, winner, status, board_state, created_at, finished_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	boardState, _ := json.Marshal(g.Board)
	var winner sql.NullString
	if g.Winner >= 0 {
		winner.String = g.Players[g.Winner].Username
		winner.Valid = true
	}

	var player2 sql.NullString
	if g.Players[1].Username != "" {
		player2.String = g.Players[1].Username
		player2.Valid = true
	}

	finishedAt := time.Now()
	if g.Status != "finished" {
		finishedAt = g.CreatedAt
	}

	_, err := s.db.Exec(query, 
		g.ID, 
		g.Players[0].Username,
		player2,
		winner,
		g.Status,
		string(boardState),
		g.CreatedAt,
		finishedAt,
	)

	if err == nil && g.Status == "finished" {
		s.updateLeaderboard(g)
	}

	return err
}

func (s *PostgresStore) updateLeaderboard(g *game.Game) {
	if g.Winner == -1 {
		// Draw - update both players
		s.updatePlayerStats(g.Players[0].Username, 0, 0, 1)
		if g.Players[1].Username != "" {
			s.updatePlayerStats(g.Players[1].Username, 0, 0, 1)
		}
	} else {
		// Winner exists
		winner := g.Players[g.Winner].Username
		loser := g.Players[1-g.Winner].Username
		
		s.updatePlayerStats(winner, 1, 0, 0)
		if !g.Players[1-g.Winner].IsBot {
			s.updatePlayerStats(loser, 0, 1, 0)
		}
	}
}

func (s *PostgresStore) updatePlayerStats(username string, wins, losses, draws int) {
	query := `
	INSERT INTO leaderboard (username, wins, losses, draws, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (username) 
	DO UPDATE SET 
		wins = leaderboard.wins + EXCLUDED.wins,
		losses = leaderboard.losses + EXCLUDED.losses,
		draws = leaderboard.draws + EXCLUDED.draws,
		updated_at = EXCLUDED.updated_at
	`

	_, err := s.db.Exec(query, username, wins, losses, draws, time.Now())
	if err != nil {
		log.Printf("Error updating leaderboard: %v", err)
	}
}

func (s *PostgresStore) GetLeaderboard() ([]LeaderboardEntry, error) {
	query := `
	SELECT username, wins, losses, draws
	FROM leaderboard 
	ORDER BY wins DESC, draws DESC, losses ASC
	LIMIT 100
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(&entry.Username, &entry.Wins, &entry.Losses, &entry.Draws)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

type LeaderboardEntry struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Draws    int    `json:"draws"`
}