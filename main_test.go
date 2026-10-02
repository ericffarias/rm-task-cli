package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
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

func TestListTasks(t *testing.T) {
	timestamp := time.Now()
	DBTest := DB{
		IdCount: 8,
		Tasks: []Task{
			{Id: 1, Description: "Tarefa concluída 1", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Tarefa concluída 2", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 3, Description: "Tarefa concluída 3", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 4, Description: "Tarefa pendente 1", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 5, Description: "Tarefa pendente 2", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 6, Description: "Tarefa pendente 3", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 7, Description: "Tarefa pendente 4", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 8, Description: "Tarefa em andamento", Status: "in-progress", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	filterByStatus := func(args []string, expect int, filename string, t *testing.T) {
		filterBy := args[2]
		rTask, err := listTasks(args, filename)
		if err != nil {
			t.Errorf("Erro no retorno da listagem de tarefas em [%v]: %v", filterBy, err)
		}
		if len(rTask) != expect {
			t.Errorf("Número de tarefas em [%v] diferente do esperado: got: %v, want: %v", filterBy, len(rTask), expect)
		}
	}
	t.Run("deve ser capaz de listar todas as tarefas", func(t *testing.T) {
		args := []string{"./task-cli", "list"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Não foi possível escrever o mock no arquivo.", err)
		}

		rTask, err := listTasks(args, filename)
		if err != nil {
			t.Error("Erro no retorno da listagem todas as tarefas:", err)
		}
		if len(rTask) != DBTest.IdCount {
			t.Errorf("Número de tarefas diferente do esperado. got: %v want: %v", len(rTask), DBTest.IdCount)
		}
	})

	t.Run("deve ser capaz de listar tarefas com status [todo]", func(t *testing.T) {
		args := []string{"./task-cli", "list", "todo"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Não foi possível escrever o mock no arquivo.", err)
		}

		filterByStatus(args, 4, filename, t)
	})
	t.Run("deve ser capaz de listar tarefas com status [done]", func(t *testing.T) {
		args := []string{"./task-cli", "list", "done"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Não foi possível escrever o mock no arquivo.", err)
		}

		filterByStatus(args, 3, filename, t)
	})
	t.Run("deve ser capaz de listar tarefas em [in-progress]", func(t *testing.T) {
		args := []string{"./task-cli", "list", "in-progress"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Não foi possível escrever o mock no arquivo.", err)
		}

		filterByStatus(args, 1, filename, t)
	})
}

func TestDeleteTask(t *testing.T) {
	timestamp := time.Now()
	DB := DB{
		IdCount: 5,
		Tasks: []Task{
			{Id: 1, Description: "Tarefa mock 1", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Tarefa mock 2", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 3, Description: "Tarefa mock 3", Status: "in-progress", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 4, Description: "Tarefa mock 4", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 5, Description: "Tarefa mock 5", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	tCount := func(filename string) int {
		db, err := readDB(filename)
		if err != nil {
			t.Errorf("Não foi possivel acessar arquivo. %v", err)
		}

		return len(db.Tasks)
	}

	t.Run("deve ser capaz de remover um id que existe", func(t *testing.T) {
		args := []string{"./task-cli", "delete", "1"}
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DB, filename)

		before := tCount(filename)

		if _, err := deleteTask(args, filename); err != nil {
			t.Error(err)
		}

		after := tCount(filename)

		if after != before-1 {
			t.Errorf("Registro não removido. got: %v, want: %v", after, before-1)
		}
	})

	t.Run("não deve remover id que não existe", func(t *testing.T) {
		args := []string{"./task-cli", "delete", "500"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DB, filename)

		before := tCount(filename)

		if _, err := deleteTask(args, filename); err == nil {
			t.Error("Um erro era esperado. Não é permitido deletar task cujo id não existe.")
		}

		after := tCount(filename)
		if before != after {
			t.Errorf("Número de task diferente. got: %v, want: %v", after, before)
		}
	})

	t.Run("não deve remover se id não fornecido", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)

		args := []string{"./task-cli", "delete"}
		saveDatabase(DB, filename)

		before := tCount(filename)

		if _, err := deleteTask(args, filename); err == nil {
			t.Error("Um erro era esperado. Não é permitido quando id não informado")
		}

		after := tCount(filename)
		if before != after {
			t.Errorf("Número de task diferente. got: %v, want: %v", after, before)
		}
	})

}
func TestUpdateTask(t *testing.T) {
	timestamp := time.Now()
	DBTest := DB{
		IdCount: 3,
		Tasks: []Task{
			{Id: 1, Description: "Primeira tarefa", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Segunda tarefa", Status: "in-progress", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 3, Description: "Terceira tarefa", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	t.Run("deve ser capaz de alterar uma task que existe", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DBTest, filename)

		newDesc := "escovar os dentes"
		args := []string{"./task-cli", "update", "2", newDesc}
		if err := updateTask(args, filename); err != nil {
			t.Error("Task não atualizada", err)
		}

		updatedDB, _ := readDB(filename)

		taskId, _ := strconv.Atoi(args[2])
		taskIndex := slices.IndexFunc(updatedDB.Tasks, func(tsk Task) bool {
			return tsk.Id == taskId
		})
		desc := updatedDB.Tasks[taskIndex].Description
		if !strings.EqualFold(desc, newDesc) {
			t.Errorf("Tarefa não foi atualizada. got: \"%v\", want: \"%v\"", desc, newDesc)
		}
	})

	t.Run("deve recusar atualizar quando descrição não informada", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DBTest, filename)

		args := []string{"./task-cli", "update", "2"}
		if err := updateTask(args, filename); err == nil {
			t.Error("Descrição não informada, um erro era esperado.")
		}
	})

	t.Run("deve recusar atualizar quando id não informado", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DBTest, filename)

		args := []string{"./task-cli", "update", "banhar os cachorros"}
		if err := updateTask(args, filename); err == nil {
			t.Error("Id não informado, um erro era esperado.")
		}
	})
}

func TestMarkTask(t *testing.T) {
	timestamp := time.Now()
	DBTest := DB{
		IdCount: 2,
		Tasks: []Task{
			{Id: 1, Description: "Primeira tarefa", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Segunda tarefa", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	assertStatusUpdated := func(t *testing.T, command, expectedStatus string) {
		t.Helper()
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Errorf("Não foi possível salvar o mock do banco: %v", err)
		}

		args := []string{"./task-cli", command, "2"}
		if err := markTask(args, filename); err != nil {
			t.Errorf("Não foi capaz de marcar a tarefa como %s: %v", expectedStatus, err)
		}

		database, _ := readDB(filename)
		taskId, _ := strconv.Atoi(args[2])
		taskIndex := slices.IndexFunc(database.Tasks, func(tsk Task) bool {
			return tsk.Id == taskId
		})

		task := &database.Tasks[taskIndex]
		if task.Status != expectedStatus {
			t.Errorf("Status da tarefa não foi atualizado. got: %v, want: %v", task.Status, expectedStatus)
		}

		if task.CreatedAt.Compare(task.UpdatedAt) != -1 {
			t.Errorf("UpdatedAt da tarefa %v não foi atualizada junto com Status.", taskId)
		}
	}

	t.Run("Deve ser capaz de trocar status para done", func(t *testing.T) {
		assertStatusUpdated(t, "mark-done", "done")
	})
	t.Run("Deve ser capaz de trocar status para in-progress", func(t *testing.T) {
		assertStatusUpdated(t, "mark-in-progress", "in-progress")
	})

	t.Run("Deve negar alterar status quando faltando operação não existir", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Fatalf("Não foi possível salvar o mock do banco: %v", err)
		}

		before, err := readDB(filename)
		if err != nil {
			t.Fatalf("Não foi possível ler o banco antes da operação inválida: %v", err)
		}

		args := []string{"./task-cli", "mark-unknown", "2"}
		if err := markTask(args, filename); err == nil {
			t.Error("markTask() deveria retornar erro quando a operação não existe")
		}

		after, err := readDB(filename)
		if err != nil {
			t.Fatalf("Não foi possível ler o banco após a operação inválida: %v", err)
		}

		taskId, _ := strconv.Atoi(args[2])
		beforeIndex := slices.IndexFunc(before.Tasks, func(tsk Task) bool { return tsk.Id == taskId })
		afterIndex := slices.IndexFunc(after.Tasks, func(tsk Task) bool { return tsk.Id == taskId })

		if before.Tasks[beforeIndex].Status != after.Tasks[afterIndex].Status {
			t.Errorf("O status da tarefa não deveria mudar para operação inválida. got: %v, want: %v", after.Tasks[afterIndex].Status, before.Tasks[beforeIndex].Status)
		}
		if !before.Tasks[beforeIndex].UpdatedAt.Equal(after.Tasks[afterIndex].UpdatedAt) {
			t.Errorf("UpdatedAt da tarefa foi alterado indevidamente para operação inválida.")
		}
	})
}
