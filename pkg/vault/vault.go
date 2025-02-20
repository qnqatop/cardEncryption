package vault

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/vault/api"
)

const (
	MaxOperations = 2 // Максимальное количество операций для DEK
)

type VaultService struct {
	api *api.Client
}

func NewVaultService(vaultAddress string) *VaultService {
	return &VaultService{
		api: New(vaultAddress),
	}
}

func New(address string) *api.Client {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Ошибка создания клиента: %v", err)
	}
	client.SetToken("myroot")

	return client
}

// GetCounter получение  счетчика из kv vault
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

// перезаписывание счетчика в vault
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

// NewDEK создание нового DEK
func (v *VaultService) NewDEK(token string) (string, error) {
	newDEK := "dek-" + token
	_, err := v.api.Logical().Write("transit/keys/"+newDEK, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка создания нового DEK: %v", err)
	}

	return newDEK, nil
}

// DecryptCard расшифрока карты с помощью ключа  dek
func (v *VaultService) DecryptCard(card, dek string) (string, error) {
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

// EncryptCard шифрование карты текущим  DEK
func (v *VaultService) EncryptCard(count int, card, dek string) (string, error) {
	// Шифрование текущим DEK
	resp, err := v.api.Logical().Write("transit/encrypt/"+dek, map[string]interface{}{
		"plaintext": base64.StdEncoding.EncodeToString([]byte(card)),
	})
	if err != nil {
		return "", err
	}
	// Увеличиваем счетчик
	log.Printf("Операция %d из %d для DEK %s\n", count, MaxOperations, dek)
	err = v.setCounter(count)
	if err != nil {
		return "", err
	}

	return resp.Data["ciphertext"].(string), nil
}
