package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"

	"github.com/sethvargo/go-password/password"
)

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func CreateNewConnectionString(ctx context.Context) (string, error) {
	id, err := randomHex(2)
	if err != nil {
		return "", err
	}

	// connect to master db
	masterAddress := os.Getenv("POSTGRES_ADDR")
	masterUser := os.Getenv("POSTGRES_USER")
	masterPassword := os.Getenv("POSTGRES_PASSWORD")
	masterPasswordUrlEncoded := url.QueryEscape(masterPassword)
	masterDb := os.Getenv("POSTGRES_DB")
	masterConnStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", masterUser, masterPasswordUrlEncoded, masterAddress, masterDb)
	db, err := sql.Open("postgres", masterConnStr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	// create new db
	newDb := "db_" + id
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", newDb))
	if err != nil {
		return "", fmt.Errorf("failed to create new database: %v", err)
	}

	// create new user
	newUser := "user_" + id
	newPassword, err := password.Generate(32, 6, 6, false, true)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %v", err)
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", newUser, newPassword))
	if err != nil {
		return "", fmt.Errorf("failed to create new user: %v", err)
	}
	_, err = db.ExecContext(ctx, fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", newDb, newUser))
	if err != nil {
		return "", fmt.Errorf("failed to set up new user privileges: %v", err)
	}

	newPasswordUrlEncoded := url.QueryEscape(newPassword)
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", newUser, newPasswordUrlEncoded, masterAddress, newDb), nil
}
