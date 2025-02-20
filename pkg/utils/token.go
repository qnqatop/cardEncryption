package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateToken() (string, error) {
	tokenBytes := make([]byte, 32)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %v", err)
	}

	token := hex.EncodeToString(tokenBytes)
	return token, nil
}
