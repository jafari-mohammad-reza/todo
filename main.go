package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/term"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Command string
type Status string

const (
	Done    Status = "done"
	Pending Status = "pending"
	Failed  Status = "failed"
)

var statusMap = map[int]Status{
	1: Done,
	2: Pending,
	3: Failed,
}

type Task struct {
	Id          int
	Title       string
	Description string
	DuDate      time.Time
	Status
	CategoryId int
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
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(dataFile, data, 0644)
}

var tasksStorage *storage[Task]
var categoryStorage *storage[Category]

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		exit()
		cancel()
	}()

	tasksStorage = newStorage[Task]("task", make(map[int]Task))
	categoryStorage = newStorage[Category]("categories", make(map[int]Category))

	command := flag.String("command", "exit", "command to enter")
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		runCommand(Command(*command))
		scanner.Scan()
		*command = scanner.Text()
	}

	<-ctx.Done()
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
		Exit:           exit,
	}
	fCmd, ok := cmdMap[cmd]
	if ok {
		fCmd()
		return
	}
	fmt.Println("unknown command")
}
func exit() {
	fmt.Println("Bye!")
	tasksStorage.backupMemoryStorage()
	categoryStorage.backupMemoryStorage()
	os.Exit(1)
}
func ListTasks() {
	tasks := tasksStorage.memoryStorage
	for t := range tasks {
		task := tasks[t]
		fmt.Printf("\n title : %s\n descritpion : %s\n status : %s\n du date : %s\n category : %s\n", task.Title, task.Description, task.Status, task.DuDate, GetCategory(task.CategoryId).Title)
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
	task.DuDate, _ = time.Parse("2006-01-02", scanner.scanInput("duDate", true, 3))
	fmt.Printf("Insert status:\nOptions:\n1)Done\n2)Failed\n3)Pending\n")
	st, _ := strconv.Atoi(scanner.scanInput("status", true, 3))
	task.Status = statusMap[st]
	fmt.Println("Insert category:\nOptions:")
	ListCategory()
	categoryId, _ := strconv.Atoi(scanner.scanInput("categoryId", true, 3))
	task.CategoryId = categoryId - 1
	tasksStorage.AddMemoryItem(task.Id, task)
	fmt.Printf("Task created successfully.\n")
	ListTasks()
}
func UpdateTask() {
	scanner := newCustomerScanner()
	var id string
	if len(os.Args) == 3 && os.Args[2] != "" {
		id = os.Args[2]
		os.Args[2] = ""
	} else {
		fmt.Println("Insert task Id:")
		id = scanner.scanInput("Id", true, -1)
	}
	Id, _ := strconv.Atoi(id)
	task := tasksStorage.memoryStorage[Id-1]
	task.Title = strings.Trim(readWithDefaultVal("Title", task.Title), " ")
	fmt.Printf("Title: %s\n", task.Title)
	task.Description = strings.Trim(readWithDefaultVal("Description", task.Description), " ")
	fmt.Printf("Description: %s\n", task.Description)
	task.Status = Status(readWithDefaultVal("Status", string(task.Status)))
	fmt.Printf("Status: %v\n", task.Status)
	tasksStorage.memoryStorage[task.Id] = task
	fmt.Println("Task updated successfully.")
	ListTasks()
}

func readWithDefaultVal(fieldName string, defaultText string) string {
	initialState, _ := term.GetState(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), initialState)
	term.MakeRaw(int(os.Stdin.Fd()))

	lineBuffer := []rune(defaultText)
	fmt.Printf("%s: %s", fieldName, defaultText)
	promptLength := len(fieldName) + 2 + len(defaultText)

	for {
		buf := make([]byte, 1)
		_, _ = os.Stdin.Read(buf)

		switch buf[0] {
		case 3: // Ctrl+C
			term.Restore(int(os.Stdin.Fd()), initialState)
			os.Exit(1)
		case 13: // Enter key
			fmt.Print("\n")
			fmt.Print("\033[1A")                 // Move cursor up one line
			fmt.Printf("\033[%dD", promptLength) // Move cursor back to the start of the prompt
			fmt.Print("\033[K")                  // Clear the line
			return string(lineBuffer)
		case 127: // Backspace
			if len(lineBuffer) > 0 {
				lineBuffer = lineBuffer[:len(lineBuffer)-1]
				fmt.Print("\b \b") // Move back, clear character, move back again
			}
		default:
			lineBuffer = append(lineBuffer, rune(buf[0]))
			fmt.Print(string(buf))
		}
	}
}

func RemoveTask() {
	var id string
	if len(os.Args) == 3 && os.Args[2] != "" {
		id = os.Args[2]
		os.Args[2] = ""
	} else {
		scanner := newCustomerScanner()
		fmt.Println("Insert task Id:")
		id = scanner.scanInput("Id", true, -1)
	}
	Id, _ := strconv.Atoi(id)
	tasksStorage.RemoveMemoryItem(Id)
}
func ListCategory() {
	for c := range categoryStorage.memoryStorage {
		category := categoryStorage.memoryStorage[c]
		fmt.Printf("\n%d)%s\n", category.Id, category.Title)
	}
}
func CreateCategory() {
	category := Category{}
	category.Id = len(categoryStorage.memoryStorage) + 1
	if len(os.Args) == 3 && os.Args[2] != "" {
		category.Title = os.Args[2]
		os.Args[2] = ""
	} else {
		scanner := newCustomerScanner()
		fmt.Println("Insert category title:")
		category.Title = scanner.scanInput("title", true, -1)
	}
	categoryStorage.AddMemoryItem(category.Id, category)
	fmt.Printf("Category created successfully.\n")
	ListCategory()
}
func UpdateCategory() {}
func RemoveCategory() {
	var id string
	if len(os.Args) == 3 && os.Args[2] != "" {
		id = os.Args[2]
		os.Args[2] = ""
	} else {
		scanner := newCustomerScanner()
		fmt.Println("Insert category Id:")
		id = scanner.scanInput("Id", true, -1)
	}
	Id, _ := strconv.Atoi(id)
	categoryStorage.RemoveMemoryItem(Id)
}
func GetCategoryByTitle(title string) *Category {
	var category Category
	for c := range categoryStorage.memoryStorage {
		if categoryStorage.memoryStorage[c].Title == title {
			category = categoryStorage.memoryStorage[c]
			return &category
		}
	}
	return nil
}
func GetCategory(id int) Category {
	if len(categoryStorage.memoryStorage) < id+1 {
		log.Fatalf("%d is invalid category", id)
	}
	return categoryStorage.memoryStorage[id]
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
