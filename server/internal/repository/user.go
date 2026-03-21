package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dz-poker/server/internal/model"
	"github.com/google/uuid"
)

// CreateUser creates a new user with initial chips
func CreateUser(username, passwordHash, nickname, avatar string) (*model.User, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := DB.Exec(
		`INSERT INTO users (id, username, password_hash, nickname, avatar, chips, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, username, passwordHash, nickname, avatar, 10000, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Create initial stats
	if err := createInitialStats(id); err != nil {
		return nil, fmt.Errorf("create initial stats: %w", err)
	}

	return &model.User{
		ID:       id,
		Username: username,
		Nickname: nickname,
		Avatar:   avatar,
		Chips:    10000,
	}, nil
}

func createInitialStats(userID string) error {
	_, err := DB.Exec(
		`INSERT INTO user_stats (user_id, total_games, wins, total_profit, created_at, updated_at)
		 VALUES (?, 0, 0, 0, ?, ?)`,
		userID, time.Now(), time.Now(),
	)
	return err
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string) (*model.User, error) {
	user := &model.User{}
	err := DB.QueryRow(
		`SELECT id, username, password_hash, nickname, avatar, chips, created_at, updated_at
		 FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.Avatar, &user.Chips, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(userID string) (*model.User, error) {
	user := &model.User{}
	err := DB.QueryRow(
		`SELECT id, username, password_hash, nickname, avatar, chips, created_at, updated_at
		 FROM users WHERE id = ?`,
		userID,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Nickname, &user.Avatar, &user.Chips, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

// GetUserStats retrieves user statistics
func GetUserStats(userID string) (*model.UserStats, error) {
	stats := &model.UserStats{}
	err := DB.QueryRow(
		`SELECT user_id, total_games, wins, total_profit, created_at, updated_at
		 FROM user_stats WHERE user_id = ?`,
		userID,
	).Scan(&stats.UserID, &stats.TotalGames, &stats.Wins, &stats.TotalProfit, &stats.CreatedAt, &stats.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user stats: %w", err)
	}
	return stats, nil
}

// UpdateUserChips updates user's chips
func UpdateUserChips(userID string, chips int64) error {
	result, err := DB.Exec(
		`UPDATE users SET chips = ?, updated_at = ? WHERE id = ?`,
		chips, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("update user chips: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// IncrementUserChips adds/subtracts chips from user
func IncrementUserChips(userID string, delta int64) error {
	_, err := DB.Exec(
		`UPDATE users SET chips = chips + ?, updated_at = ? WHERE id = ?`,
		delta, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("increment user chips: %w", err)
	}
	return nil
}

// UpdateUserStats updates user statistics after a game
func UpdateUserStats(userID string, profit int64, won bool) error {
	wins := 0
	if won {
		wins = 1
	}
	_, err := DB.Exec(
		`UPDATE user_stats
		 SET total_games = total_games + ?,
		     wins = wins + ?,
		     total_profit = total_profit + ?,
		     updated_at = ?
		 WHERE user_id = ?`,
		1, wins, profit, time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("update user stats: %w", err)
	}
	return nil
}

// GetUserWithStats retrieves user with statistics
func GetUserWithStats(userID string) (*model.User, *model.UserStats, error) {
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, nil
	}

	stats, err := GetUserStats(userID)
	if err != nil {
		return nil, nil, err
	}

	return user, stats, nil
}

// ValidateAvatar checks if avatar is valid
func ValidateAvatar(avatar string) bool {
	for _, a := range model.Avatars {
		if a == avatar {
			return true
		}
	}
	return false
}

// GetAllUsers retrieves all users (for admin)
func GetAllUsers() ([]*model.User, error) {
	rows, err := DB.Query(
		`SELECT id, username, nickname, avatar, chips, created_at, updated_at FROM users ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("get all users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Nickname, &user.Avatar, &user.Chips, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}
