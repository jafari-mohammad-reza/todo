package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	buildCmd := exec.Command("go", "build", "-o", "todo")
	if err := buildCmd.Run(); err != nil {
		os.Exit(1)
	}
	exitCode := m.Run()
	os.Remove("todo")
	os.Exit(exitCode)
}

func TestCreateCategoryWithArg(t *testing.T) {
	os.Args = []string{"./todo", "create-category", "first-category"}
	categoryStorage = newStorage[Category]("categories", make(map[int]Category))
	CreateCategory()
	if len(categoryStorage.memoryStorage) == 0 {
		t.Error("Category 'first-category' not created")
	}
}
