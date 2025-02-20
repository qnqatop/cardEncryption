package cardstorage

import (
	"fmt"
	"log"

	"encryptCard/pkg/db"
	"encryptCard/pkg/utils"
	"encryptCard/pkg/vault"
)

type Service struct {
	CardStorage  *db.CardRepository
	DekStorage   *db.DEKRepository
	vault        *vault.VaultService
	currentDEK   string
	maxOperation int
	count        int
}

func New(vaultAddress, dek string, maxOperation int) *Service {
	return &Service{
		CardStorage:  db.NewCardRepository(),
		DekStorage:   db.NewDEKRepository(),
		vault:        vault.NewVaultService(vaultAddress),
		currentDEK:   dek,
		maxOperation: maxOperation,
	}
}

// Шифрование с DEK и учет счетчика
func (cs *Service) EncryptWithDEK(card string) (string, error) {
	count, err := cs.vault.GetCounter()
	if err != nil {
		return "", fmt.Errorf("ошибка получения счетчика: %v", err)
	}

	log.Printf("Текущий счетчик в vault - %d", count)
	log.Printf("Текущий счетчик в sorage - %d", cs.count)

	// Проверка лимита и ротация DEK
	if cs.count >= cs.maxOperation {
		log.Println("Счетчик DEK исчерпан, ротация ключа...")
		err = cs.RotateDEK()
		if err != nil {
			return "", fmt.Errorf("ошибка ротации DEK: %v", err)
		}
		cs.count = 0

	}

	cs.count++
	encCard, err := cs.vault.EncryptCard(cs.count, card, cs.currentDEK)
	if err != nil {
		cs.count -= 1
		return "", err
	}

	cardToken, err := utils.GenerateToken()
	if err != nil {
		return "", err
	}

	nc := db.EncryptCard{
		EcnCard: encCard,
		Token:   cardToken,
		Dec:     cs.currentDEK,
	}

	cs.CardStorage.Add(nc)

	return encCard, nil
}

// Ротация DEK
func (cs *Service) RotateDEK() error {
	// Создаем новый DEK
	// генерируем новое имя
	t, err := utils.GenerateToken()
	if err != nil {
		return err
	}
	newDEK, err := cs.vault.NewDEK(t)

	// фиксируем старый dek для дешифровки
	oldDEK := cs.currentDEK

	// делаем новый основной dek для шифрования
	cs.currentDEK = newDEK

	for _, v := range cs.CardStorage.GetAll() {
		if v.Dec == oldDEK {
			data, err := cs.DecryptCard(v.EcnCard, oldDEK)
			if err != nil {
				return err
			}

			v.EcnCard, err = cs.EncryptWithDEK(data)
			if err != nil {
				return err
			}

			log.Printf("обновили карту -- `%s` -- \nc DEK %s\nна %s", data, oldDEK, cs.currentDEK)
			cs.CardStorage.Add(v)
		}
	}

	// Обновляем текущий DEK
	log.Printf("DEK ротирован, новый ключ: %s\n", cs.currentDEK)
	log.Printf("Старый ключ - %s\n", oldDEK)
	return nil
}

func (cs *Service) DecryptCard(encCard, dek string) (string, error) {
	if dek == "" {
		dek = cs.currentDEK
	}
	data, err := cs.vault.DecryptCard(encCard, dek)
	if err != nil {
		return "", err
	}

	return data, nil
}
