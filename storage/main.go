package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Storage[T any] interface {
	SaveItem(id int, item T)
	RemoveItem(id int)
	GetItem(id int) (*T, error)
	GetItems() []T
	Close() error
}
type MemoryStorage[T any] struct {
	name string
	data map[int]T
}

func NewMemoryStorage[T any](name string) *MemoryStorage[T] {
	s := &MemoryStorage[T]{
		name: name,
		data: make(map[int]T),
	}
	s.loadMemoryStorage()
	return s
}

func (s *MemoryStorage[T]) loadMemoryStorage() error {
	config, _ := os.UserConfigDir()
	dataFile := fmt.Sprintf("%s/todo/%s.json", config, s.name)
	data, err := os.ReadFile(dataFile)
	if os.IsNotExist(err) {
		err := os.WriteFile(dataFile, []byte("[]"), 0644)
		if err != nil {
			return err
		}
	} else {
		if err != nil {
			return err
		}
		var datas []T
		err = json.Unmarshal(data, &datas)
		if err != nil {
			return err
		}
		for i, v := range datas {
			s.data[i] = v
		}
	}
	return nil
}
func (s *MemoryStorage[T]) backupMemoryStorage() error {
	config, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dataFile := fmt.Sprintf("%s/todo/%s.json", config, s.name)
	var datas []T
	for _, v := range s.data {
		datas = append(datas, v)
	}
	data, err := json.Marshal(datas)
	if err != nil {
		return err
	}
	err = os.WriteFile(dataFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *MemoryStorage[T]) Close() error {
	return s.backupMemoryStorage()
}

func (s *MemoryStorage[T]) SaveItem(id int, item T) {
	s.data[id] = item
}

func (s *MemoryStorage[T]) RemoveItem(id int) {
	delete(s.data, id)
}
func (s *MemoryStorage[T]) GetItem(id int) (*T, error) {
	item, exist := s.data[id]
	if !exist {
		return nil, errors.New("item not found")
	}
	return &item, nil
}

func (s *MemoryStorage[T]) GetItems() []T {
	items := make([]T, 0, len(s.data))
	for _, item := range s.data {
		items = append(items, item)
	}
	return items
}
