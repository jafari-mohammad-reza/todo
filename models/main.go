package models

import "time"

type Repository[T any] interface {
	Get(id int) (*T, error)
	List() []T
	Save(model T) error
	Delete(id int) error
	CloseStorage() error
}

type BaseModel struct {
	Id        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
