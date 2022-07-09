package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
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
	id, err := randomHex(6)
	if err != nil {
		return "", err
	}

	// connect to master db
	masterAddress := os.Getenv("POSTGRES_ADDR")
	masterUser := os.Getenv("POSTGRES_USER")
	masterPassword := os.Getenv("POSTGRES_PASSWORD")
	masterDb := os.Getenv("POSTGRES_DB")
	masterConnStr := fmt.Sprintf("postgres://%s:%s@%s/%s", masterUser, masterPassword, masterAddress, masterDb)
	db, err := sql.Open("postgres", masterConnStr)
	if err != nil {
		return "", err
	}
	defer db.Close()

	// create new db
	newDb := "db-" + id
	_, err = db.ExecContext(ctx, "CREATE DATABASE %s;", newDb)
	if err != nil {
		return "", err
	}

	// create new user
	newUser := "user-" + id
	newPassword, err := password.Generate(32, 6, 6, false, true)
	if err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx, "CREATE USER $1 WITH PASSWORD '$2';", newUser, newPassword)
	if err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx, "GRANT ALL PRIVILEGES ON DATABASE $1 TO $2;", newDb, newUser)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgres://%s:%s@%s/%s", newUser, newPassword, masterAddress, newDb), nil
}
