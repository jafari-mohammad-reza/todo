package models

import (
	"errors"
	"fmt"
	"time"
	"todo-cli/storage"
)

type Status string

const (
	Done    Status = "done"
	Pending Status = "pending"
	Failed  Status = "failed"
)

var StatusMap = map[int]Status{
	1: Done,
	2: Pending,
	3: Failed,
}

type Task struct {
	BaseModel
	Title       string
	Description string
	DuDate      time.Time
	Status
	CategoryId int
}
type TaskRepository struct {
	storage storage.Storage[Task]
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		storage: storage.NewMemoryStorage[Task]("tasks"),
	}
}

func (t TaskRepository) Get(id int) (*Task, error) {
	task, err := t.storage.GetItem(id)
	if err != nil {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (t TaskRepository) List() []Task {
	return t.storage.GetItems()
}

func (t TaskRepository) Save(model Task) error {
	model.Id = len(t.List()) + 1
	err := t.storage.SaveItem(model.Id, model)
	if err != nil {
		return fmt.Errorf("failed to create task. %s", err.Error())
	}
	return nil
}

func (t TaskRepository) Delete(id int) error {
	_, err := t.Get(id)
	if err != nil {
		return err
	}
	err = t.storage.RemoveItem(id)
	if err != nil {
		return fmt.Errorf("failed to delete task. %s", err.Error())
	}
	return nil
}

func (t TaskRepository) CloseStorage() error {
	return t.storage.Close()
}
