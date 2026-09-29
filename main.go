package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
			fmt.Println("Não foi possível salvar a task.")
			fmt.Println(err.Error())
		}
		return
	case "list":
		tasks, err := listTasks(args, FILE_NAME)
		if err != nil {
			fmt.Println("Erro na operação de filteragem:", err.Error())
			return
		}

		fmt.Printf("%v registros encontrados.\n", len(tasks))
		fmt.Printf("\t  Id Estatus \t Descrição\n")
		for _, t := range tasks {
			fmt.Printf("\t- %02d [%v] \t %v\n", t.Id, t.Status, t.Description)
		}
		return
	}

}

func addTask(args []string, filename string) error {
	var newTask Task

	if len(args) < 3 {
		return errors.New("Descrição não encontrada.")
	}

	description := args[2]

	if description == "" {
		return errors.New("Tentativa de salvar Task com descrição vazia.")
	}

	db, err := readDB(filename)
	if err != nil {
		fmt.Println("Não foi possível acessar as tarefas")
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
	database, err := readDB(filename)

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
		return nil, fmt.Errorf("Operção de listagem por \"%v\" não encontrada", listBy)
	}

	filteredList := filter(database.Tasks, func(t Task) bool {
		return strings.EqualFold(t.Status, listBy)
	})

	return filteredList, nil
}

func saveDatabase(db DB, filename string) error {
	bytesToWrite, err := json.Marshal(db)
	if err != nil {
		fmt.Println("Não foi possível os dados.")
		return err
	}

	if err := os.WriteFile(filename, bytesToWrite, PERMISSIONS); err != nil {
		fmt.Println("Não foi possível salvar o arquivo.")
		return err
	}

	return nil
}

func printHelp() {
	fmt.Println("===================== Task Cli ======================")
	fmt.Println("Uso: task-cli <comando> [argumentos]")
	fmt.Println("\nComandos disponíveis:")
	fmt.Println("  add <descrição>              Cria uma nova tarefa")
	fmt.Println("  update <id> <descrição>      Atualiza o texto de uma tarefa existente")
	fmt.Println("  delete <id>                  Remove a tarefa <id>")
	fmt.Println("  mark-in-progress <id>        Marca a tarefa <id> como \"in-progress\"")
	fmt.Println("  mark-done <id>               Marca a tarefa <id> como \"done\"")
	fmt.Println("  list                         Lista as tarefas")
	fmt.Println("  list	done                   Lista as tarefas marcadas como \"done\"")
	fmt.Println("  list	in-progress            Lista as tarefas marcadas como \"in-progress\"")
	fmt.Println("  list	todo                   Lista as tarefas marcadas como \"todo\"")
	fmt.Println(".")
	fmt.Println(".")
	fmt.Println(".")
}

func readDB(filename string) (DB, error) {
	var database DB

	file, errFile := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, PERMISSIONS)
	if errFile != nil {
		fmt.Printf("Erro ao brir o arquivo.")
		return database, errFile
	}
	defer file.Close()

	blob, errBlob := io.ReadAll(file)
	if errBlob != nil {
		fmt.Printf("Erro na leitura do arquivo.")
		return database, errBlob
	}

	if len(blob) < 1 {
		blob = []byte("{}")
		if _, err := file.Write(blob); err != nil {
			return database, err
		}
	}

	err := json.Unmarshal(blob, &database)
	if err != nil {
		fmt.Println("Erro na converção na descerialização.")
		return database, err
	}

	return database, nil
}
