package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

type Command string
type Status string

const (
	Done    Status = "done"
	Pending Status = "pending"
	Failed  Status = "failed"
)

type Task struct {
	Id          int
	Title       string
	Description string
	DuDate      string
	Status
	*Category
}

type Category struct {
	Id    int
	Title string
}

const (
	createTask     Command = "create-task"
	createCategory Command = "create-category"
	listTasks      Command = "list-tasks"
	listCategory   Command = "list-category"
	removeTask     Command = "remove-task"
	removeCategory Command = "remove-category"
	updateTask     Command = "update-task" // in this case we will print each field of task in stdout and let user change it and pass it again as stdin
	updateCategory Command = "update-category"
	Exit           Command = "exit"
)

type customerScanner struct {
	scanner *bufio.Scanner
}

func newCustomerScanner() *customerScanner {
	return &customerScanner{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

type storage[T any] struct {
	name          string
	memoryStorage map[int]T
}

func newStorage[T any](name string, memoryStorage map[int]T) *storage[T] {
	s := &storage[T]{
		name:          name,
		memoryStorage: memoryStorage,
	}
	s.loadMemoryStorage()
	return s
}

func (s *storage[T]) AddMemoryItem(id int, item T) {
	s.memoryStorage[id] = item
}

func (s *storage[T]) RemoveMemoryItem(id int) {
	delete(s.memoryStorage, id)
}
func (s *storage[T]) loadMemoryStorage() {
	config, _ := os.UserConfigDir()
	dataFile := fmt.Sprintf("%s/todo/%s.json", config, s.name)
	data, err := os.ReadFile(dataFile)
	if os.IsNotExist(err) {
		err := os.WriteFile(dataFile, []byte("[]"), 0644)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		if err != nil {
			log.Fatal(err)
		}
		var datas []T
		err = json.Unmarshal(data, &datas)
		if err != nil {
			log.Fatal(err)
		}
		for i, v := range datas {
			s.memoryStorage[i] = v
		}
	}
}
func (s *storage[T]) backupMemoryStorage() {
	config, _ := os.UserConfigDir()
	dataFile := fmt.Sprintf("%s/todo/%s.json", config, s.name)
	var datas []T
	for _, v := range s.memoryStorage {
		datas = append(datas, v)
	}
	data, err := json.Marshal(datas)
	fmt.Println(data)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(dataFile, data, 0644)
}

var tasksStorage *storage[Task]
var categoryStorage *storage[Category]

func main() {
	tasksStorage = newStorage[Task]("task", make(map[int]Task))
	categoryStorage = newStorage[Category]("categories", make(map[int]Category))
	command := flag.String("command", "exit", "command to enter")
	flag.Parse()
	for {
		runCommand(Command(*command))
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		*command = scanner.Text()
	}
}

func runCommand(cmd Command) {
	cmdMap := map[Command]func(){
		listTasks:      ListTasks,
		createTask:     CreateTask,
		updateTask:     UpdateTask,
		removeTask:     RemoveTask,
		listCategory:   ListCategory,
		createCategory: CreateCategory,
		updateCategory: UpdateCategory,
		removeCategory: RemoveCategory,
		Exit: func() {
			fmt.Println("Bye!")
			tasksStorage.backupMemoryStorage()
			categoryStorage.backupMemoryStorage()
			os.Exit(1)
		},
	}
	fCmd, ok := cmdMap[cmd]
	if ok {
		fCmd()
		return
	}
	fmt.Println("unknown command")
}
func ListTasks() {
	tasks := tasksStorage.memoryStorage
	for t := range tasks {
		task := tasks[t]
		category := "unknown"
		if task.Category != nil {
			category = task.Category.Title
		}
		fmt.Printf(" title : %s\n descritpion : %s\n status : %s\n du date : %s\n category : %s\n", task.Title, task.Description, task.Status, task.DuDate, category)
	}
}
func CreateTask() {
	task := Task{}
	task.Id = len(tasksStorage.memoryStorage) + 1
	scanner := newCustomerScanner()
	fmt.Println("Insert task title:")
	task.Title = scanner.scanInput("title", true, 3)
	fmt.Println("Insert task description:")
	task.Description = scanner.scanInput("description", true, 3)
	fmt.Println("Insert duDate:")
	task.DuDate = scanner.scanInput("duDate", true, 3)
	fmt.Println("Insert status:")
	task.Status = Status(scanner.scanInput("status", true, 3))
	fmt.Println("Insert category:")
	task.Category = GetCategory(scanner.scanInput("category", true, 3))
	tasksStorage.AddMemoryItem(task.Id, task)
	fmt.Printf("Task created successfully.\n")
	ListTasks()
}
func UpdateTask() {}
func RemoveTask() {}
func ListCategory() {
	for c := range categoryStorage.memoryStorage {
		fmt.Printf("\n%+v\n", categoryStorage.memoryStorage[c])
	}
}
func CreateCategory() {
	scanner := newCustomerScanner()
	category := Category{}
	category.Id = len(categoryStorage.memoryStorage) + 1
	if len(os.Args) == 3 && os.Args[2] != "" {
		category.Title = os.Args[2]
		os.Args[2] = ""
	} else {
		fmt.Println("Insert category title:")
		category.Title = scanner.scanInput("title", true, -1)
	}
	categoryStorage.AddMemoryItem(category.Id, category)
	fmt.Printf("Category created successfully.\n")
	ListCategory()
}
func UpdateCategory() {}
func RemoveCategory() {}
func GetCategory(title string) *Category {
	var category Category
	for c := range categoryStorage.memoryStorage {
		if categoryStorage.memoryStorage[c].Title == title {
			category = categoryStorage.memoryStorage[c]
			return &category
		}
	}
	return nil
}
func (c *customerScanner) scanInput(title string, required bool, maxTry int) string {
	// maxTry -1 means infinite loop
	tryCount := 0
	for {
		if maxTry != -1 {
			if tryCount > maxTry {
				log.Fatalf("no value entered for required filed: %s\n", title)
				return ""
			}
			tryCount++
		}
		c.scanner.Scan()
		scanned := c.scanner.Text()
		if required && scanned == "" {
			fmt.Printf("\nPlease enter a valid input for %s.\n", title)
		} else {
			return scanned
		}
	}
}
