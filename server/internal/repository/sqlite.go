package repository

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes SQLite database and runs migrations
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = runMigrations(); err != nil {
		return err
	}

	log.Printf("Database initialized at %s", dbPath)
	return nil
}

func runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			nickname TEXT NOT NULL,
			avatar TEXT NOT NULL,
			chips INT64 DEFAULT 10000,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id TEXT PRIMARY KEY,
			total_games INT64 DEFAULT 0,
			wins INT64 DEFAULT 0,
			total_profit INT64 DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS game_history (
			id TEXT PRIMARY KEY,
			room_code TEXT NOT NULL,
			user_id TEXT NOT NULL,
			initial_chips INT64 NOT NULL,
			final_chips INT64 NOT NULL,
			profit INT64 NOT NULL,
			final_rank INT NOT NULL,
			best_hand TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_game_history_user_id ON game_history(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_game_history_room_code ON game_history(room_code)`,
	}

	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			log.Printf("Migration failed: %s, error: %v", m, err)
			return err
		}
	}

	return nil
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
