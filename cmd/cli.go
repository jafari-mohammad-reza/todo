package cmd

import (
	"bufio"
	"context"
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
	"todo-cli/models"
)

type Command string

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

var taskRepository models.Repository[models.Task]
var categoryRepository models.Repository[models.Category]

func RunCli() {
	taskRepository = models.NewTaskRepository()
	categoryRepository = models.NewCategoryRepository()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM) // listen for sigterm to execute exit() method and backup all memory data.

	go func() {
		<-signalChan
		exit()
		cancel()
	}()

	os.Args[1] = os.Args[2] // to replace "cli" in beginning
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
	taskRepository.CloseStorage()
	categoryRepository.CloseStorage()
	os.Exit(1)
}
func ListTasks() {
	tasks := taskRepository.List()
	for t := range tasks {
		task := tasks[t]
		fmt.Printf("\n title : %s\n descritpion : %s\n status : %s\n du date : %s\n category : %s\n", task.Title, task.Description, task.Status, task.DuDate, GetCategory(task.CategoryId).Title)
	}
}
func CreateTask() {
	task := models.Task{}
	scanner := newCustomerScanner()
	fmt.Println("Insert task title:")
	task.Title = scanner.scanInput("title", true, 3)
	fmt.Println("Insert task description:")
	task.Description = scanner.scanInput("description", true, 3)
	fmt.Println("Insert duDate:")
	task.DuDate, _ = time.Parse("2006-01-02", scanner.scanInput("duDate", true, 3))
	fmt.Printf("Insert status:\nOptions:\n1)Done\n2)Failed\n3)Pending\n")
	st, _ := strconv.Atoi(scanner.scanInput("status", true, 3))
	task.Status = models.StatusMap[st]
	fmt.Println("Insert category:\nOptions:")
	ListCategory()
	categoryId, _ := strconv.Atoi(scanner.scanInput("categoryId", true, 3))
	task.CategoryId = categoryId - 1
	taskRepository.Save(task)
	fmt.Printf("Task created successfully.\n")
	ListTasks()
}
func UpdateTask() {
	scanner := newCustomerScanner()
	var id string
	if len(os.Args) == 4 && os.Args[3] != "" {
		id = os.Args[3]
		os.Args[3] = ""
	} else {
		fmt.Println("Insert task Id:")
		id = scanner.scanInput("Id", true, -1)
	}
	Id, _ := strconv.Atoi(id)
	task, err := taskRepository.Get(Id - 1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	task.Title = strings.Trim(readWithDefaultVal("Title", task.Title, true), " ")
	fmt.Printf("Title: %s\n", task.Title)
	task.Description = strings.Trim(readWithDefaultVal("Description", task.Description, true), " ")
	fmt.Printf("Description: %s\n", task.Description)
	task.Status = models.Status(readWithDefaultVal("Status", string(task.Status), true))
	fmt.Printf("Status: %v\n", task.Status)
	taskRepository.Save(*task)
	fmt.Println("Task updated successfully.")
	ListTasks()
}

func readWithDefaultVal(fieldName string, defaultText string, required bool) string {
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
			if required && string(lineBuffer) == "" {
				fmt.Printf("%s can not be empty.", fieldName)
				readWithDefaultVal(fieldName, defaultText, required)
			}
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
	if len(os.Args) == 4 && os.Args[3] != "" {
		id = os.Args[3]
		os.Args[3] = ""
	} else {
		scanner := newCustomerScanner()
		fmt.Println("Insert task Id:")
		id = scanner.scanInput("Id", true, -1)
	}
	Id, _ := strconv.Atoi(id)
	taskRepository.Delete(Id)
}
func ListCategory() {
	for _, c := range categoryRepository.List() {
		fmt.Printf("\n%d)%s\n", c.Id, c.Title)
	}
}
func CreateCategory() {
	category := models.Category{}
	if len(os.Args) == 4 && os.Args[3] != "" {
		category.Title = os.Args[3]
		os.Args[3] = ""
	} else {
		scanner := newCustomerScanner()
		fmt.Println("Insert category title:")
		category.Title = scanner.scanInput("title", true, -1)
	}
	categoryRepository.Save(category)
	fmt.Printf("Category created successfully.\n")
	ListCategory()
}
func UpdateCategory() {
	scanner := newCustomerScanner()
	var id string
	if len(os.Args) == 4 && os.Args[3] != "" {
		id = os.Args[3]
		os.Args[3] = ""
	} else {
		fmt.Println("Insert category Id:")
		id = scanner.scanInput("Id", true, -1)
	}
	Id, _ := strconv.Atoi(id)
	category, err := categoryRepository.Get(Id - 1)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	category.Title = strings.Trim(readWithDefaultVal("Title", category.Title, true), " ")
	fmt.Printf("Title: %s\n", category.Title)
	categoryRepository.Save(*category)
	fmt.Println("category updated successfully.")
	ListCategory()
}
func RemoveCategory() {
	var id string
	if len(os.Args) == 4 && os.Args[3] != "" {
		id = os.Args[3]
		os.Args[3] = ""
	} else {
		scanner := newCustomerScanner()
		fmt.Println("Insert category Id:")
		id = scanner.scanInput("Id", true, -1)
	}
	Id, _ := strconv.Atoi(id)
	categoryRepository.Delete(Id)
}

func GetCategory(id int) models.Category {
	category, err := categoryRepository.Get(id)
	if err != nil {
		log.Fatal(err.Error())
	}
	return *category
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
