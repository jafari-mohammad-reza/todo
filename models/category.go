package models

import (
	"errors"
	"fmt"
	"todo-cli/storage"
)

type Category struct {
	BaseModel
	Title string
}
type CategoryRepository struct {
	storage storage.Storage[Category]
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		storage: storage.NewMemoryStorage[Category]("categories"),
	}
}

func (r CategoryRepository) Get(id int) (*Category, error) {
	category, err := r.storage.GetItem(id)
	if err != nil {
		return nil, errors.New("category not found")
	}
	return category, nil
}

func (r CategoryRepository) List() []Category {
	return r.storage.GetItems()
}

func (r CategoryRepository) Save(model Category) error {
	model.Id = len(r.List()) + 1
	err := r.storage.SaveItem(model.Id, model)
	if err != nil {
		return fmt.Errorf("failed to create category. %s", err.Error())
	}
	return nil
}
func (r CategoryRepository) Delete(id int) error {
	_, err := r.Get(id)
	if err != nil {
		return err
	}
	err = r.storage.RemoveItem(id)
	if err != nil {
		return fmt.Errorf("failed to delete category. %s", err.Error())
	}
	return nil
}

func (r CategoryRepository) CloseStorage() error {
	return r.storage.Close()
}
