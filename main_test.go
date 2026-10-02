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
	t.Run("should be able to save and read DB data from a .json file", func(t *testing.T) {
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
			t.Fatalf("Attempt to save DB failed: %v", err)
		}

		savedData, err := readDatabase(tempFileUrl)
		if err != nil {
			t.Fatalf("Attempt to read DB failed: %v", err)
		}
		if savedData.IdCount != dataToSave.IdCount {
			t.Errorf("Different ID counter: got %d, want %d", savedData.IdCount, dataToSave.IdCount)
		}
		if len(savedData.Tasks) != len(dataToSave.Tasks) {
			t.Fatalf("Different number of tasks: got %d, want %d", len(savedData.Tasks), len(dataToSave.Tasks))
		}

		got, want := savedData.Tasks[0], dataToSave.Tasks[0]
		if got.Id != want.Id || got.Description != want.Description || got.Status != want.Status ||
			!got.CreatedAt.Equal(want.CreatedAt) || !got.UpdatedAt.Equal(want.UpdatedAt) {
			t.Errorf("Saved task data differ: got %v, want %v", got, want)
		}
	})

	t.Run("should initialize an empty file", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := os.WriteFile(filename, nil, PERMISSIONS); err != nil {
			t.Fatalf("Could not create empty file: %v", err)
		}

		database, err := readDatabase(filename)
		if err != nil {
			t.Fatalf("readDatabase() failed for empty file: %v", err)
		}
		if database.IdCount != 0 || len(database.Tasks) != 0 {
			t.Errorf("Database initialized differently than expected: got %+v", database)
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Could not read the initialized file: %v", err)
		}
		if !json.Valid(content) {
			t.Errorf("The initialized content is not valid JSON: %q", content)
		}
	})

	t.Run("should return an error for invalid JSON", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := os.WriteFile(filename, []byte("{"), PERMISSIONS); err != nil {
			t.Fatalf("Could not create invalid file: %v", err)
		}
		if _, err := readDatabase(filename); err == nil {
			t.Error("readDatabase() should return an error for invalid JSON")
		}
	})
}

func TestAddTask(t *testing.T) {
	t.Run("should be able to create a task", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		descriptions := []string{"First task", "Second task"}
		for _, description := range descriptions {
			args := []string{"task-cli", "add", description}
			if err := addTask(args, filename); err != nil {
				t.Fatalf("addTask() failed: %v", err)
			}
		}

		database, err := readDatabase(filename)
		if err != nil {
			t.Fatalf("Could not read the provided file: %v", err)
		}
		if database.IdCount != 2 {
			t.Errorf("Incorrect ID counter: got %d, want 2", database.IdCount)
		}
		if len(database.Tasks) != len(descriptions) {
			t.Fatalf("Incorrect number of tasks: got %d, want %d", len(database.Tasks), len(descriptions))
		}
		for index, task := range database.Tasks {
			if task.Id != index+1 {
				t.Errorf("Incorrect ID in task %d: got %d, want %d", index, task.Id, index+1)
			}
			if task.Description != descriptions[index] {
				t.Errorf("Incorrect description in task %d: got %q, want %q", index, task.Description, descriptions[index])
			}
			if task.Status != "todo" {
				t.Errorf("Incorrect status in task %d: got %q, want %q", index, task.Status, "todo")
			}
			if task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
				t.Errorf("Timestamps were not filled in task %d: %+v", index, task)
			}
		}
	})

	t.Run("should reject blank description without changing the database", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DB{IdCount: 3}, filename); err != nil {
			t.Fatalf("Could not prepare the database: %v", err)
		}
		args := []string{"task-cli", "add", ""}

		if err := addTask(args, filename); err == nil {
			t.Error("addTask() should reject empty descriptions")
		}

		database, err := readDatabase(filename)
		if err != nil {
			t.Fatalf("Could not read the database after rejecting the description: %v", err)
		}
		if database.IdCount != 3 || len(database.Tasks) != 0 {
			t.Errorf("The database changed: got %+v, want idCount 3 and no tasks", database)
		}
	})

	t.Run("should reject arguments without description", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := addTask([]string{"task-cli", "add"}, filename); err == nil {
			t.Error("addTask() should return an error when description is not provided")
		}
		if _, err := os.Stat(filename); !os.IsNotExist(err) {
			t.Errorf("The file should not be created for invalid arguments; stat error: %v", err)
		}
	})

	t.Run("should return an error when it cannot save", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), "nonexistent-directory", FILE_NAME)
		if err := saveDatabase(DB{}, filename); err == nil {
			t.Error("saveDatabase() should return an error if the directory does not exist")
		}
	})
}

func TestListTasks(t *testing.T) {
	timestamp := time.Now()
	DBTest := DB{
		IdCount: 8,
		Tasks: []Task{
			{Id: 1, Description: "Completed task 1", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Completed task 2", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 3, Description: "Completed task 3", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 4, Description: "Pending task 1", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 5, Description: "Pending task 2", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 6, Description: "Pending task 3", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 7, Description: "Pending task 4", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 8, Description: "In progress task", Status: "in-progress", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	filterByStatus := func(args []string, expect int, filename string, t *testing.T) {
		filterBy := args[2]
		rTask, err := listTasks(args, filename)
		if err != nil {
			t.Errorf("Error in task list return for [%v]: %v", filterBy, err)
		}
		if len(rTask) != expect {
			t.Errorf("Incorrect number of tasks in [%v]: got: %v, want: %v", filterBy, len(rTask), expect)
		}
	}
	t.Run("should be able to list all tasks", func(t *testing.T) {
		args := []string{"./task-cli", "list"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Could not write the mock file.", err)
		}

		rTask, err := listTasks(args, filename)
		if err != nil {
			t.Error("Error in the list all tasks return:", err)
		}
		if len(rTask) != DBTest.IdCount {
			t.Errorf("Incorrect number of tasks. got: %v want: %v", len(rTask), DBTest.IdCount)
		}
	})

	t.Run("should be able to list tasks with status [todo]", func(t *testing.T) {
		args := []string{"./task-cli", "list", "todo"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Could not write the mock file.", err)
		}

		filterByStatus(args, 4, filename, t)
	})
	t.Run("should be able to list tasks with status [done]", func(t *testing.T) {
		args := []string{"./task-cli", "list", "done"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Could not write the mock file.", err)
		}

		filterByStatus(args, 3, filename, t)
	})
	t.Run("should be able to list tasks in [in-progress]", func(t *testing.T) {
		args := []string{"./task-cli", "list", "in-progress"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Error("Could not write the mock file.", err)
		}

		filterByStatus(args, 1, filename, t)
	})
}

func TestDeleteTask(t *testing.T) {
	timestamp := time.Now()
	DB := DB{
		IdCount: 5,
		Tasks: []Task{
			{Id: 1, Description: "Mock task 1", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Mock task 2", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 3, Description: "Mock task 3", Status: "in-progress", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 4, Description: "Mock task 4", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 5, Description: "Mock task 5", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	tCount := func(filename string) int {
		db, err := readDatabase(filename)
		if err != nil {
			t.Errorf("Could not access the file. %v", err)
		}

		return len(db.Tasks)
	}

	t.Run("should be able to remove an existing id", func(t *testing.T) {
		args := []string{"./task-cli", "delete", "1"}
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DB, filename)

		before := tCount(filename)

		if _, err := deleteTask(args, filename); err != nil {
			t.Error(err)
		}

		after := tCount(filename)

		if after != before-1 {
			t.Errorf("Record was not removed. got: %v, want: %v", after, before-1)
		}
	})

	t.Run("should not remove a non-existent id", func(t *testing.T) {
		args := []string{"./task-cli", "delete", "500"}

		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DB, filename)

		before := tCount(filename)

		if _, err := deleteTask(args, filename); err == nil {
			t.Error("An error was expected. Deleting a task whose id does not exist is not allowed.")
		}

		after := tCount(filename)
		if before != after {
			t.Errorf("Number of tasks different. got: %v, want: %v", after, before)
		}
	})

	t.Run("should not remove when id is not provided", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)

		args := []string{"./task-cli", "delete"}
		saveDatabase(DB, filename)

		before := tCount(filename)

		if _, err := deleteTask(args, filename); err == nil {
			t.Error("An error was expected. This is not allowed when id is not provided")
		}

		after := tCount(filename)
		if before != after {
			t.Errorf("Number of tasks different. got: %v, want: %v", after, before)
		}
	})

}
func TestUpdateTask(t *testing.T) {
	timestamp := time.Now()
	DBTest := DB{
		IdCount: 3,
		Tasks: []Task{
			{Id: 1, Description: "First task", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Second task", Status: "in-progress", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 3, Description: "Third task", Status: "done", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	t.Run("should be able to change a task that exists", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DBTest, filename)

		newDesc := "brush your teeth"
		args := []string{"./task-cli", "update", "2", newDesc}
		if err := updateTask(args, filename); err != nil {
			t.Error("Task not updated", err)
		}

		updatedDB, _ := readDatabase(filename)

		taskId, _ := strconv.Atoi(args[2])
		taskIndex := slices.IndexFunc(updatedDB.Tasks, func(tsk Task) bool {
			return tsk.Id == taskId
		})
		desc := updatedDB.Tasks[taskIndex].Description
		if !strings.EqualFold(desc, newDesc) {
			t.Errorf("Task was not updated. got: \"%v\", want: \"%v\"", desc, newDesc)
		}
	})

	t.Run("should reject updating when description is not provided", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DBTest, filename)

		args := []string{"./task-cli", "update", "2"}
		if err := updateTask(args, filename); err == nil {
			t.Error("Description not provided, an error was expected.")
		}
	})

	t.Run("should reject updating when id is not provided", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		saveDatabase(DBTest, filename)

		args := []string{"./task-cli", "update", "walk the dogs"}
		if err := updateTask(args, filename); err == nil {
			t.Error("Id not provided, an error was expected.")
		}
	})
}

func TestMarkTask(t *testing.T) {
	timestamp := time.Now()
	DBTest := DB{
		IdCount: 2,
		Tasks: []Task{
			{Id: 1, Description: "First task", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
			{Id: 2, Description: "Second task", Status: "todo", CreatedAt: timestamp, UpdatedAt: timestamp},
		},
	}

	assertStatusUpdated := func(t *testing.T, command, expectedStatus string) {
		t.Helper()
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Errorf("Could not save the database mock: %v", err)
		}

		args := []string{"./task-cli", command, "2"}
		if err := markTask(args, filename); err != nil {
			t.Errorf("Could not mark the task as %s: %v", expectedStatus, err)
		}

		database, _ := readDatabase(filename)
		taskId, _ := strconv.Atoi(args[2])
		taskIndex := slices.IndexFunc(database.Tasks, func(tsk Task) bool {
			return tsk.Id == taskId
		})

		task := &database.Tasks[taskIndex]
		if task.Status != expectedStatus {
			t.Errorf("Task status was not updated. got: %v, want: %v", task.Status, expectedStatus)
		}

		if task.CreatedAt.Compare(task.UpdatedAt) != -1 {
			t.Errorf("UpdatedAt for task %v was not updated along with Status.", taskId)
		}
	}

	t.Run("should be able to change status to done", func(t *testing.T) {
		assertStatusUpdated(t, "mark-done", "done")
	})
	t.Run("should be able to change status to in-progress", func(t *testing.T) {
		assertStatusUpdated(t, "mark-in-progress", "in-progress")
	})

	t.Run("should reject changing status when the operation does not exist", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), FILE_NAME)
		if err := saveDatabase(DBTest, filename); err != nil {
			t.Fatalf("Could not save the database mock: %v", err)
		}

		before, err := readDatabase(filename)
		if err != nil {
			t.Fatalf("Could not read the database before the invalid operation: %v", err)
		}

		args := []string{"./task-cli", "mark-unknown", "2"}
		if err := markTask(args, filename); err == nil {
			t.Error("markTask() should return an error when the operation does not exist")
		}

		after, err := readDatabase(filename)
		if err != nil {
			t.Fatalf("Could not read the database after the invalid operation: %v", err)
		}

		taskId, _ := strconv.Atoi(args[2])
		beforeIndex := slices.IndexFunc(before.Tasks, func(tsk Task) bool { return tsk.Id == taskId })
		afterIndex := slices.IndexFunc(after.Tasks, func(tsk Task) bool { return tsk.Id == taskId })

		if before.Tasks[beforeIndex].Status != after.Tasks[afterIndex].Status {
			t.Errorf("Task status should not change for an invalid operation. got: %v, want: %v", after.Tasks[afterIndex].Status, before.Tasks[beforeIndex].Status)
		}
		if !before.Tasks[beforeIndex].UpdatedAt.Equal(after.Tasks[afterIndex].UpdatedAt) {
			t.Errorf("UpdatedAt for the task was changed improperly for an invalid operation.")
		}
	})
}
