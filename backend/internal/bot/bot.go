package bot

import (
	"connect-four/internal/game"
	"math/rand"
	"time"
)

type Bot struct {
	Difficulty int
}

func NewBot() *Bot {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())
	return &Bot{
		Difficulty: 2, // Medium difficulty
	}
}

func (b *Bot) CalculateMove(g *game.Game) int {
	// Strategy: 
	// 1. Check if bot can win immediately
	// 2. Check if need to block opponent
	// 3. Create winning opportunities
	// 4. Prefer center columns

	botPlayer := 2 // Assuming bot is always player 2
	humanPlayer := 1

	// Check for immediate win
	for col := 0; col < 7; col++ {
		if b.isValidMove(g, col) {
			row := b.getNextAvailableRow(g, col)
			if row != -1 {
				// Simulate move
				g.Board[row][col] = botPlayer
				if g.CheckWin(row, col) {
					g.Board[row][col] = 0 // Reset
					return col
				}
				g.Board[row][col] = 0 // Reset
			}
		}
	}

	// Block opponent's immediate win
	for col := 0; col < 7; col++ {
		if b.isValidMove(g, col) {
			row := b.getNextAvailableRow(g, col)
			if row != -1 {
				// Simulate opponent move
				g.Board[row][col] = humanPlayer
				if g.CheckWin(row, col) {
					g.Board[row][col] = 0 // Reset
					return col
				}
				g.Board[row][col] = 0 // Reset
			}
		}
	}

	// Create opportunities (3 in a row with potential for 4)
	priorityMoves := []int{}
	for col := 0; col < 7; col++ {
		if b.isValidMove(g, col) {
			row := b.getNextAvailableRow(g, col)
			if row != -1 {
				g.Board[row][col] = botPlayer
				if b.hasPotentialWin(g, botPlayer, 3) {
					priorityMoves = append(priorityMoves, col)
				}
				g.Board[row][col] = 0
			}
		}
	}

	if len(priorityMoves) > 0 {
		return priorityMoves[rand.Intn(len(priorityMoves))]
	}

	// Create opportunities (2 in a row with potential for 4)
	secondaryMoves := []int{}
	for col := 0; col < 7; col++ {
		if b.isValidMove(g, col) {
			row := b.getNextAvailableRow(g, col)
			if row != -1 {
				g.Board[row][col] = botPlayer
				if b.hasPotentialWin(g, botPlayer, 2) {
					secondaryMoves = append(secondaryMoves, col)
				}
				g.Board[row][col] = 0
			}
		}
	}

	if len(secondaryMoves) > 0 {
		return secondaryMoves[rand.Intn(len(secondaryMoves))]
	}

	// Prefer center columns for better positioning
	centerPref := []int{3, 2, 4, 1, 5, 0, 6}
	for _, col := range centerPref {
		if b.isValidMove(g, col) {
			return col
		}
	}

	// Fallback: random valid move (should always find one if board isn't full)
	validMoves := []int{}
	for col := 0; col < 7; col++ {
		if b.isValidMove(g, col) {
			validMoves = append(validMoves, col)
		}
	}

	if len(validMoves) > 0 {
		return validMoves[rand.Intn(len(validMoves))]
	}

	return -1 // Should never reach here if board isn't full
}

func (b *Bot) isValidMove(g *game.Game, column int) bool {
	if column < 0 || column >= 7 {
		return false
	}
	return g.Board[0][column] == 0
}

func (b *Bot) getNextAvailableRow(g *game.Game, column int) int {
	for row := 5; row >= 0; row-- {
		if g.Board[row][column] == 0 {
			return row
		}
	}
	return -1
}

func (b *Bot) hasPotentialWin(g *game.Game, player int, minSequence int) bool {
	// Check for sequences that could lead to wins
	for row := 0; row < 6; row++ {
		for col := 0; col < 7; col++ {
			if g.Board[row][col] == player || g.Board[row][col] == 0 {
				// Check horizontal potential
				if b.checkDirectionPotential(g, player, row, col, 0, 1) >= minSequence {
					return true
				}
				// Check vertical potential
				if b.checkDirectionPotential(g, player, row, col, 1, 0) >= minSequence {
					return true
				}
				// Check diagonal potentials
				if b.checkDirectionPotential(g, player, row, col, 1, 1) >= minSequence {
					return true
				}
				if b.checkDirectionPotential(g, player, row, col, 1, -1) >= minSequence {
					return true
				}
			}
		}
	}
	return false
}

func (b *Bot) checkDirectionPotential(g *game.Game, player, startRow, startCol, rowDelta, colDelta int) int {
	count := 0
	consecutive := true
	
	for i := 0; i < 4; i++ {
		row := startRow + i*rowDelta
		col := startCol + i*colDelta
		
		if row < 0 || row >= 6 || col < 0 || col >= 7 {
			break
		}
		
		// Count if it's our piece or empty (potential)
		if g.Board[row][col] == player {
			count++
		} else if g.Board[row][col] == 0 && consecutive {
			// Only count empty spaces if we haven't broken the sequence
			count++
		} else {
			// Opponent piece or broken sequence
			consecutive = false
			break
		}
	}
	return count
}

// Additional helper function for better strategy
func (b *Bot) evaluatePosition(g *game.Game, player int) int {
	score := 0
	
	// Center column preference
	for row := 0; row < 6; row++ {
		if g.Board[row][3] == player {
			score += 3
		}
	}
	
	// Horizontal opportunities
	for row := 0; row < 6; row++ {
		for col := 0; col < 4; col++ {
			window := [4]int{
				g.Board[row][col],
				g.Board[row][col+1],
				g.Board[row][col+2],
				g.Board[row][col+3],
			}
			score += b.evaluateWindow(window, player)
		}
	}
	
	return score
}

func (b *Bot) evaluateWindow(window [4]int, player int) int {
	score := 0
	opponent := 3 - player // Since players are 1 and 2
	
	playerCount := 0
	opponentCount := 0
	emptyCount := 0
	
	for _, cell := range window {
		if cell == player {
			playerCount++
		} else if cell == opponent {
			opponentCount++
		} else {
			emptyCount++
		}
	}
	
	// Score based on the window configuration
	if playerCount == 4 {
		score += 100
	} else if playerCount == 3 && emptyCount == 1 {
		score += 5
	} else if playerCount == 2 && emptyCount == 2 {
		score += 2
	}
	
	if opponentCount == 3 && emptyCount == 1 {
		score -= 4 // Block opponent
	}
	
	return score
}