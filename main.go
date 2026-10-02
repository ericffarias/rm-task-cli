package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	Id          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type DB struct {
	IdCount int    `json:"idCount"`
	Tasks   []Task `json:"tasks"`
}

const FILE_NAME = "data.json"
const PERMISSIONS = 0600

func main() {
	Run(os.Args)
}

func Run(args []string) {

	if len(args) < 2 {
		printHelp()
		return
	}

	action := args[1]

	switch action {
	case "add":
		if err := addTask(args, FILE_NAME); err != nil {
			fmt.Println("Could not save the task.")
			fmt.Println(err.Error())
		}
	case "list":
		tasks, err := listTasks(args, FILE_NAME)
		if err != nil {
			fmt.Println("Error in filtering operation:", err.Error())
			return
		}

		fmt.Printf("%v records found.\n", len(tasks))
		fmt.Printf("\t  Id Status \t Description\n")

		for _, t := range tasks {
			fmt.Printf("\t- %02d [%v] \t %v\n", t.Id, t.Status, t.Description)
		}
	case "delete":
		if task, err := deleteTask(args, FILE_NAME); err != nil {
			fmt.Println("Error while trying to remove:", err)
		} else {
			fmt.Printf("task [%v %v] removed.\n", task.Id, task.Description)
		}
	case "update":
		if err := updateTask(args, FILE_NAME); err != nil {
			fmt.Println("Update error:", err)
		}
	case "mark-in-progress", "mark-done":
		if err := markTask(args, FILE_NAME); err != nil {
			fmt.Println("Error updating status:", err)
		}
	default:
		fmt.Println("Operation not recognized")
	}

}

func markTask(args []string, filename string) error {
	if len(args) < 3 {
		return errors.New("Operation and id are required.")
	}

	operation := args[1]
	taskId, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	database, err := readDatabase(filename)
	if err != nil {
		return err
	}

	taskIndex := slices.IndexFunc(database.Tasks, func(t Task) bool {
		return t.Id == taskId
	})
	task := &database.Tasks[taskIndex]

	switch operation {
	case "mark-in-progress":
		task.Status = "in-progress"
	case "mark-done":
		task.Status = "done"
	default:
		return fmt.Errorf("Operation %v not recognized.", operation)
	}

	task.UpdatedAt = time.Now()

	if err := saveDatabase(database, filename); err != nil {
		return err
	}

	return nil
}

func updateTask(args []string, filename string) error {
	if len(args) < 4 {
		return errors.New("To update a task, you must provide the id and the new description.")
	}

	description := args[3]
	id, err := strconv.Atoi(args[2])
	if err != nil {
		return err
	}

	database, err := readDatabase(filename)
	if err != nil {
		return err
	}

	taskIdx := slices.IndexFunc(database.Tasks, func(t Task) bool {
		return t.Id == id
	})

	nTask := &database.Tasks[taskIdx]
	nTask.Description = description
	nTask.UpdatedAt = time.Now()

	if err := saveDatabase(database, filename); err != nil {
		return err
	}

	return nil
}

func deleteTask(args []string, filename string) (Task, error) {
	var removedTask Task

	if len(args) < 3 {
		return removedTask, errors.New("Removal id not provided.")
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		return removedTask, fmt.Errorf("Provided id must be an integer %v", err)
	}

	database, err := readDatabase(filename)
	if err != nil {
		return removedTask, err
	}

	if i := slices.IndexFunc(database.Tasks, func(t Task) bool {
		return t.Id == id
	}); i >= 0 {
		removedTask = database.Tasks[i]
	} else {
		return removedTask, fmt.Errorf("Id %v not found.", id)
	}

	database.Tasks = slices.DeleteFunc(database.Tasks, func(t Task) bool {
		return t.Id == id
	})

	if err := saveDatabase(database, filename); err != nil {
		return removedTask, err
	}

	return removedTask, nil
}

func addTask(args []string, filename string) error {
	var newTask Task

	if len(args) < 3 {
		return errors.New("Description not found.")
	}

	description := args[2]

	if description == "" {
		return errors.New("Attempt to save a task with an empty description.")
	}

	db, err := readDatabase(filename)
	if err != nil {
		fmt.Println("Unable to access tasks")
		return err
	}

	db.IdCount++
	timestamp := time.Now()
	newTask = Task{
		Id:          db.IdCount,
		Description: description,
		Status:      "todo",
		CreatedAt:   timestamp,
		UpdatedAt:   timestamp,
	}

	db.Tasks = append(db.Tasks, newTask)
	if err := saveDatabase(db, filename); err != nil {
		return err
	}

	return nil
}

func filter[T any](s []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}

	return result
}

func listTasks(args []string, filename string) ([]Task, error) {
	database, err := readDatabase(filename)

	if err != nil {
		return nil, err
	}

	if len(args) < 3 {
		return database.Tasks, nil
	}

	listBy := args[2]

	var isValid bool = false
	for _, v := range []string{"done", "todo", "in-progress"} {
		if strings.EqualFold(listBy, v) {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, fmt.Errorf("List operation by \"%v\" not found", listBy)
	}

	filteredList := filter(database.Tasks, func(t Task) bool {
		return strings.EqualFold(t.Status, listBy)
	})

	return filteredList, nil
}

func saveDatabase(db DB, filename string) error {
	bytesToWrite, err := json.Marshal(db)
	if err != nil {
		fmt.Println("Could not process data.")
		return err
	}

	if err := os.WriteFile(filename, bytesToWrite, PERMISSIONS); err != nil {
		fmt.Println("Could not save the file.")
		return err
	}

	return nil
}

func printHelp() {
	fmt.Println("===================== Task Cli ======================")
	fmt.Println("Usage: task-cli <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  add <description>             Creates a new task")
	fmt.Println("  update <id> <description>     Updates the text of an existing task")
	fmt.Println("  delete <id>                   Removes task <id>")
	fmt.Println("  mark-in-progress <id>         Marks task <id> as \"in-progress\"")
	fmt.Println("  mark-done <id>                Marks task <id> as \"done\"")
	fmt.Println("  list                          Lists tasks")
	fmt.Println("  list	done                    Lists tasks marked as \"done\"")
	fmt.Println("  list	in-progress             Lists tasks marked as \"in-progress\"")
	fmt.Println("  list	todo                    Lists tasks marked as \"todo\"")
}

func readDatabase(filename string) (DB, error) {
	var database DB

	file, errFile := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, PERMISSIONS)
	if errFile != nil {
		return database, fmt.Errorf("Error opening file: %v", errFile)
	}
	defer file.Close()

	blob, errBlob := io.ReadAll(file)
	if errBlob != nil {
		return database, fmt.Errorf("Error reading file: %v", errBlob)
	}

	if len(blob) < 1 {
		blob = []byte("{}")
		if _, err := file.Write(blob); err != nil {
			return database, err
		}
	}

	err := json.Unmarshal(blob, &database)
	if err != nil {
		return database, fmt.Errorf("Conversion/deserialization error: %v", err)
	}

	return database, nil
}
