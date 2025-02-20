package vault

import (
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/vault/api"
	"log"
	"strconv"
)

const (
	MaxOperations = 2 // Максимальное количество операций для DEK
)

type VaultService struct {
	api *api.Client
}

func NewVaultService() *VaultService {
	return &VaultService{
		api: New(),
	}
}

func New() *api.Client {
	config := api.DefaultConfig()
	config.Address = "http://localhost:8200"
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Ошибка создания клиента: %v", err)
	}
	client.SetToken("myroot")

	return client
}

// InitKV Включение KV Engine для хранения счетчика
func (v *VaultService) InitKV() error {
	_, err := v.api.Logical().Write("sys/mounts/secret", map[string]interface{}{
		"type":    "kv",
		"options": map[string]string{"version": "2"},
	})
	if err != nil {
		log.Printf("KV уже включён или ошибка: %v", err)
	}
	return err
}

func (v *VaultService) GetCounter() (int, error) {
	secret, err := v.api.Logical().Read("secret/data/dek-counter")
	if err != nil || secret == nil || secret.Data["data"] == nil {
		return 0, nil // Если счетчика нет, начинаем с 0
	}
	data := secret.Data["data"].(map[string]interface{})
	countStr, ok := data["count"].(string)
	if !ok {
		return 0, nil
	}
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (v *VaultService) setCounter(count int) error {
	_, err := v.api.Logical().Write("secret/data/dek-counter", map[string]interface{}{
		"data": map[string]interface{}{
			"count": strconv.Itoa(count),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (v *VaultService) NewDEK(token string) (string, error) {
	newDEK := "dek-" + token
	_, err := v.api.Logical().Write("transit/keys/"+newDEK, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания нового DEK: %v", err)
	}

	return newDEK, nil
}

func (v *VaultService) DecryptWithToken(card, dek string) (string, error) {
	resp, err := v.api.Logical().Write("transit/decrypt/"+dek, map[string]interface{}{
		"ciphertext": card,
	})
	if err != nil {
		return "", err
	}
	plaintextBase64 := resp.Data["plaintext"].(string)
	plaintext, err := base64.StdEncoding.DecodeString(plaintextBase64)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (v *VaultService) EncryptWithToken(count int, card, dek string) (string, error) {
	// Шифрование текущим DEK
	resp, err := v.api.Logical().Write("transit/encrypt/"+dek, map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString([]byte(card)),
	})
	if err != nil {
		return "", err
	}
	// Увеличиваем счетчик
	count++
	log.Printf("Операция %d из %d для DEK %s\n", count, MaxOperations, dek)
	err = v.setCounter(count)
	if err != nil {
		return "", err
	}

	return resp.Data["ciphertext"].(string), nil
}
