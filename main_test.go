package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDatabase(t *testing.T) {
	t.Run("deve ser capaz de salvar e ler dados do tipo DB de um arquivo .json", func(t *testing.T) {
		timestamp := time.Now()
		dataToSave := DB{
			IdCount: 1,
			Tasks: []Task{{
				Id:          1,
				Description: "Test save on database",
				Status:      "done",
				CreatedAt:   timestamp,
				UpdatedAt:   timestamp,
			}},
		}
		tempFileUrl := filepath.Join(t.TempDir(), FILE_NAME)

		if err := saveDatabase(dataToSave, tempFileUrl); err != nil {
			t.Fatalf("Tentativa de salvar DB falhou: %v", err)
		}

		savedData, err := readDB(tempFileUrl)
		if err != nil {
			t.Fatalf("Tentativa de ler DB falhou: %v", err)
		}
		if savedData.IdCount != dataToSave.IdCount {
			t.Errorf("Contador de IDs diferente: got %d, want %d", savedData.IdCount, dataToSave.IdCount)
		}
		if len(savedData.Tasks) != len(dataToSave.Tasks) {
			t.Fatalf("Quantidade de tarefas diferente: got %d, want %d", len(savedData.Tasks), len(dataToSave.Tasks))
		}

		got, want := savedData.Tasks[0], dataToSave.Tasks[0]
		if got.Id != want.Id || got.Description != want.Description || got.Status != want.Status ||
			!got.CreatedAt.Equal(want.CreatedAt) || !got.UpdatedAt.Equal(want.UpdatedAt) {
			t.Errorf("Dados da tarefa salvos diferentes: got %v, want %v", got, want)
		}
	})

	t.Run("deve inicializar um arquivo vazio", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := os.WriteFile(filename, nil, PERMISSIONS); err != nil {
			t.Fatalf("Não foi possível criar arquivo vazio: %v", err)
		}

		database, err := readDB(filename)
		if err != nil {
			t.Fatalf("readDB() falhou para arquivo vazio: %v", err)
		}
		if database.IdCount != 0 || len(database.Tasks) != 0 {
			t.Errorf("Banco inicializado diferente do esperado: got %+v", database)
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Não foi possível ler o arquivo inicializado: %v", err)
		}
		if !json.Valid(content) {
			t.Errorf("O conteúdo inicializado não é JSON válido: %q", content)
		}
	})

	t.Run("deve retornar erro para JSON inválido", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := os.WriteFile(filename, []byte("{"), PERMISSIONS); err != nil {
			t.Fatalf("Não foi possível criar arquivo inválido: %v", err)
		}
		if _, err := readDB(filename); err == nil {
			t.Error("readDB() deveria retornar erro para JSON inválido")
		}
	})
}

func TestAddTask(t *testing.T) {
	t.Run("deve ser capaz de criar uma tarefa", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		descriptions := []string{"Primeira tarefa", "Segunda tarefa"}
		for _, description := range descriptions {
			args := []string{"task-cli", "add", description}
			if err := addTask(args, filename); err != nil {
				t.Fatalf("addTask() falhou: %v", err)
			}
		}

		database, err := readDB(filename)
		if err != nil {
			t.Fatalf("Não foi possível ler o arquivo informado: %v", err)
		}
		if database.IdCount != 2 {
			t.Errorf("Contador de IDs incorreto: got %d, want 2", database.IdCount)
		}
		if len(database.Tasks) != len(descriptions) {
			t.Fatalf("Quantidade de tarefas incorreta: got %d, want %d", len(database.Tasks), len(descriptions))
		}
		for index, task := range database.Tasks {
			if task.Id != index+1 {
				t.Errorf("ID incorreto na tarefa %d: got %d, want %d", index, task.Id, index+1)
			}
			if task.Description != descriptions[index] {
				t.Errorf("Descrição incorreta na tarefa %d: got %q, want %q", index, task.Description, descriptions[index])
			}
			if task.Status != "todo" {
				t.Errorf("Status incorreto na tarefa %d: got %q, want %q", index, task.Status, "todo")
			}
			if task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
				t.Errorf("Timestamps não foram preenchidos na tarefa %d: %+v", index, task)
			}
		}
	})

	t.Run("deve rejeitar descrição em branco sem alterar o banco", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DB{IdCount: 3}, filename); err != nil {
			t.Fatalf("Não foi possível preparar o banco: %v", err)
		}
		args := []string{"task-cli", "add", ""}

		if err := addTask(args, filename); err == nil {
			t.Error("addTask() deveria rejeitar descrição vazia")
		}

		database, err := readDB(filename)
		if err != nil {
			t.Fatalf("Não foi possível ler o banco após rejeitar a descrição: %v", err)
		}
		if database.IdCount != 3 || len(database.Tasks) != 0 {
			t.Errorf("O banco foi alterado: got %+v, want idCount 3 e nenhuma tarefa", database)
		}
	})

	t.Run("deve rejeitar argumentos sem descrição", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := addTask([]string{"task-cli", "add"}, filename); err == nil {
			t.Error("addTask() deveria retornar erro quando a descrição não é informada")
		}
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			t.Errorf("O arquivo não deveria ser criado para argumentos inválidos; stat error: %v", err)
		}
	})

	t.Run("deve retornar erro quando não consegue salvar", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), "diretorio-inexistente", FILE_NAME)
		if err := saveDatabase(DB{}, filename); err == nil {
			t.Error("saveDatabase() deveria retornar erro se o diretório não existir")
		}
	})
}
