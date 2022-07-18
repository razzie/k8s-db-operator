package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"

	"github.com/mediocregopher/radix/v4"
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
	masterAddress := os.Getenv("REDIS_ADDR")
	masterPassword := os.Getenv("REDIS_PASSWORD")
	conn, err := radix.Dial(ctx, "tcp", masterAddress)
	if err != nil {
		return "", fmt.Errorf("failed to connect to redis: %v", err)
	}
	defer conn.Close()
	err = conn.Do(ctx, radix.Cmd(nil, "AUTH", masterPassword))
	if err != nil {
		return "", fmt.Errorf("redis auth failed: %v", err)
	}

	// new namespace
	newNs := "ns_" + id
	newPassword, err := password.Generate(32, 6, 6, false, true)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %v", err)
	}
	err = conn.Do(ctx, radix.Cmd(nil, "NAMESPACE", "ADD", newNs, newPassword))
	if err != nil {
		return "", fmt.Errorf("failed to create new redis namespace: %v", err)
	}

	newPasswordUrlEncoded := url.QueryEscape(newPassword)
	return fmt.Sprintf("redis://:%s@%s", newPasswordUrlEncoded, masterAddress), nil
}
