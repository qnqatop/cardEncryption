package db

import "sync"

type CardRepository struct {
	mu             *sync.RWMutex
	CardRepository map[string]EncryptCard
}

func NewCardRepository() *CardRepository {
	return &CardRepository{
		mu:             new(sync.RWMutex),
		CardRepository: make(map[string]EncryptCard),
	}
}

func (c *CardRepository) Add(card EncryptCard) {
	c.mu.Lock()
	c.CardRepository[card.Token] = card
	c.mu.Unlock()
}

func (c *CardRepository) Get(token string) (EncryptCard, bool) {
	c.mu.RLock()
	card, ok := c.CardRepository[token]
	c.mu.RUnlock()
	return card, ok
}

func (c *CardRepository) GetAll() map[string]EncryptCard {
	return c.CardRepository
}

type DEKRepository struct {
	mu            *sync.RWMutex
	DEKRepository map[string]Dek
}

func NewDEKRepository() *DEKRepository {
	return &DEKRepository{
		mu:            new(sync.RWMutex),
		DEKRepository: make(map[string]Dek),
	}
}

func (d *DEKRepository) Add(dek Dek) {
	d.mu.Lock()
	d.DEKRepository[dek.Key] = dek
	d.mu.Unlock()
}

func (d *DEKRepository) Get(key string) (Dek, bool) {
	d.mu.Lock()
	dek, ok := d.DEKRepository[key]
	d.mu.RUnlock()
	return dek, ok
}

func (d *DEKRepository) GetAll() map[string]Dek {
	return d.DEKRepository
}
