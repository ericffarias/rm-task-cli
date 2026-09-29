package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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
