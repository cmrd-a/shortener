package storage

import (
	"errors"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type InMemoryRepository struct {
	store map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{store: make(map[string]string)}
}

func (a InMemoryRepository) Get(id string) (link string, err error) {
	link, ok := a.store[id]
	if !ok {
		return "", errors.New("link not found")
	}
	return link, nil
}

func (a InMemoryRepository) Add(link string) (id string, err error) {
	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	id = string(b)
	a.store[id] = link
	return id, nil
}
