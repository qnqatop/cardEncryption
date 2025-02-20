package main

import (
	"fmt"
	"log"

	"encryptCard/pkg/cardStorage"
	"encryptCard/pkg/utils"
)

const (
	MaxOperations = 2 // Максимальное количество операций для DEK поставил мало для проверки перезаписывания
)

func main() {
	// Текущий DEK
	// Пока что при инициализации постоянно новый но! потом будем сохранять в базу и брать последний из базы
	currentDEK, err := utils.GenerateToken()
	if err != nil {
		log.Fatal(err)
	}

	// инициализация cardStorage
	cs := cardStorage.New("http://localhost:8200", currentDEK, MaxOperations)

	// Пример шифрования нескольких карт
	cards := []string{"1234-5678-9012-3456", "9876-5432-1098-7654", "1111-2222-3333-4444"}

	for _, card := range cards {
		log.Println("-------------------------------------")
		log.Printf("Шифруем карту - %s", card)
		log.Println("-------------------------------------")

		encrypted, err := cs.EncryptWithDEK(card)
		if err != nil {
			log.Fatalf("Ошибка шифрования: %v", err)
		}
		log.Printf("Зашифровано %s\n", encrypted)

		data, err := cs.DecryptCard(encrypted, "")
		if err != nil {
			log.Fatalf("Ошибка: %v", err)
		}
		log.Println("-------------------------------------")

		log.Printf("Расшифровали карту - %s\n", data)
		log.Println("-------------------------------------")
		fmt.Println("\n\n\n")
	}
}
