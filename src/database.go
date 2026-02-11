package src

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL non définie")
	}

	u, err := url.Parse(dbURL)
	if err != nil {
		return nil, err
	}

	password, _ := u.User.Password()

	dsn := fmt.Sprintf("%s:%s@tcp(%s)%s?parseTime=true&charset=utf8mb4&loc=Local",
		u.User.Username(),
		password,
		u.Host,
		u.Path,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	DB = db
	return DB, nil
}

func Migrate(db *sql.DB) error {
	const usersTable = `
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    pseudo VARCHAR(255) DEFAULT NULL,
    bio TEXT DEFAULT NULL,
    photo_profil VARCHAR(500) DEFAULT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`

	if _, err := db.Exec(usersTable); err != nil {
		return fmt.Errorf("création table users: %w", err)
	}

	alterQueries := []string{
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS pseudo VARCHAR(255) DEFAULT NULL",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS bio TEXT DEFAULT NULL",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS photo_profil VARCHAR(500) DEFAULT NULL",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user'",
	}

	for _, query := range alterQueries {
		_, _ = db.Exec(query)
	}

	var columnExists int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'role'").Scan(&columnExists)
	if err == nil && columnExists == 0 {
		_, _ = db.Exec("ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user'")
	}

	return nil
}
