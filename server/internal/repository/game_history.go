package repository

import (
	"fmt"
	"time"

	"github.com/dz-poker/server/internal/model"
	"github.com/google/uuid"
)

// RecordGameResult records a single player's game result
func RecordGameResult(roomCode, userID string, initialChips, finalChips int64, rank int, bestHand string) error {
	id := uuid.New().String()
	profit := finalChips - initialChips

	_, err := DB.Exec(
		`INSERT INTO game_history (id, room_code, user_id, initial_chips, final_chips, profit, final_rank, best_hand, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, roomCode, userID, initialChips, finalChips, profit, rank, bestHand, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("record game result: %w", err)
	}
	return nil
}

// GetUserGameHistory retrieves game history for a user
func GetUserGameHistory(userID string, limit int) ([]model.GameHistory, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := DB.Query(
		`SELECT id, room_code, user_id, initial_chips, final_chips, profit, final_rank, best_hand, created_at
		 FROM game_history WHERE user_id = ?
		 ORDER BY created_at DESC LIMIT ?`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get user game history: %w", err)
	}
	defer rows.Close()

	var histories []model.GameHistory
	for rows.Next() {
		var h model.GameHistory
		err := rows.Scan(&h.ID, &h.RoomCode, &h.UserID, &h.InitialChips, &h.FinalChips, &h.Profit, &h.FinalRank, &h.BestHand, &h.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan game history: %w", err)
		}
		histories = append(histories, h)
	}

	return histories, nil
}

// GetUserGameHistoryCount returns total number of games for a user
func GetUserGameHistoryCount(userID string) (int, error) {
	var count int
	err := DB.QueryRow(
		`SELECT COUNT(*) FROM game_history WHERE user_id = ?`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get game history count: %w", err)
	}
	return count, nil
}
