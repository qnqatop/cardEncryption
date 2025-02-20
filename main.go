package main

import (
	"crypto/rand"
	"encoding/hex"
	"encryptCard/vault"
	"fmt"
	"log"
)

const (
	MaxOperations = 2 // Максимальное количество операций для DEK
)

var (
	currentDEK = "my-dek" // Текущий DEK

)

type CardStorageService struct {
	CardStorage map[string]EncryptCard
	DekStorage  map[string]Dek
	vault       *vault.VaultService
}

func NewCardStorageService() *CardStorageService {
	return &CardStorageService{
		CardStorage: make(map[string]EncryptCard),
		DekStorage:  make(map[string]Dek),
		vault:       vault.NewVaultService(),
	}
}

func main() {
	// инициализация cardStorage
	cs := NewCardStorageService()

	// Логика шифрования карты

	// Пример шифрования нескольких карт
	cards := []string{"1234-5678-9012-3456", "9876-5432-1098-7654", "1111-2222-3333-4444"}

	for _, card := range cards {
		encrypted, err := cs.EncryptWithDEK(card)
		if err != nil {
			log.Fatalf("Ошибка шифрования: %v", err)
		}
		log.Printf("Зашифровано (%s): %s\n", currentDEK, encrypted)
	}
}

// Шифрование с DEK и учет счетчика
func (cs *CardStorageService) EncryptWithDEK(card string) (string, error) {
	// Получаем текущий счетчик
	count, err := cs.vault.GetCounter()
	if err != nil {
		return "", fmt.Errorf("ошибка получения счетчика: %v", err)
	}

	log.Printf("Текущий счетчик - %d", count)

	// Проверка лимита и ротация DEK
	if count >= MaxOperations {
		log.Println("Счетчик DEK исчерпан, ротация ключа...")
		err = cs.RotateDEK()
		if err != nil {
			return "", fmt.Errorf("ошибка ротации DEK: %v", err)
		}
		count = 0 // Сбрасываем после ротации
	}

	encCard, err := cs.vault.EncryptWithToken(count, card, currentDEK)
	if err != nil {
		return "", err
	}

	cardToken, err := generateToken()
	if err != nil {
		return "", err
	}

	cs.CardStorage[cardToken] = EncryptCard{
		EcnCard: encCard,
		Token:   cardToken,
		Dec:     currentDEK,
	}

	return encCard, nil
}

// Получение счетчика из vault

// Ротация DEK
func (cs *CardStorageService) RotateDEK() error {
	// Создаем новый DEK
	// генерируем новое имя
	t, err := generateToken()
	if err != nil {
		return err
	}
	newDEK, err := cs.vault.NewDEK(t)

	// фиксируем старый dek для дешифровки
	oldDEK := currentDEK

	// делаем новый основной dek для шифрования
	currentDEK = newDEK

	for _, v := range cs.CardStorage {
		if v.Dec == oldDEK {
			data, err := cs.vault.DecryptWithToken(v.EcnCard, oldDEK)
			if err != nil {
				return err
			}

			v.EcnCard, err = cs.EncryptWithDEK(data)
			if err != nil {
				return err
			}

			log.Printf("обновили карту %s c DEK %s на %s", data, oldDEK, currentDEK)
			cs.CardStorage[v.Token] = v
		}
	}

	// Обновляем текущий DEK
	log.Printf("DEK ротирован, новый ключ: %s\n", currentDEK)
	return nil
}

func generateToken() (string, error) {
	tokenBytes := make([]byte, 32)

	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации токена: %v", err)
	}

	// Форматируем как hex-строку (64 символа)
	// Можно также использовать base64: base64.StdEncoding.EncodeToString(tokenBytes)
	token := hex.EncodeToString(tokenBytes)
	return token, nil
}
