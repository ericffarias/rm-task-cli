# Task CLI

A simple command-line task manager. This project practices handling positional arguments, working with the file system, and persisting data in JSON without external libraries.

The application lets you create, edit, delete, and update the status of tasks, as well as list all tasks or filter them by status. Each task has an ID, description, status, creation timestamp, and last-updated timestamp. Data is stored in `data.json` in the current directory; the file is created automatically when needed.

Roadmap Project Detail: https://roadmap.sh/projects/task-tracker

## Requirements

- Go 1.27.1 or later
- Git to download the source code

## Download and build

Clone the repository and change to the project directory:

```sh
git clone https://github.com/ericffarias/task-cli.git
cd task-cli
```

Build the executable:

```sh
go build -o task-cli .
```

The executable will be created in the current directory. Run it with `./task-cli` (Linux/macOS) or `task-cli.exe` (Windows). The examples below use Linux/macOS.

## Usage

Run commands from the directory where you want to keep `data.json`.

### Add a task

```sh
./task-cli add "Buy groceries"
```

New tasks start with the `todo` status and receive an ID, creation timestamp, and update timestamp.

### Update a task description

```sh
./task-cli update 1 "Buy groceries and cook dinner"
```

Use the task ID to select which task to update. Quote the new description if it contains spaces.

### Delete a task

```sh
./task-cli delete 1
```

### Mark a task as in progress

```sh
./task-cli mark-in-progress 1
```

### Mark a task as done

```sh
./task-cli mark-done 1
```

### List tasks

List all tasks:

```sh
./task-cli list
```

Filter tasks by status:

```sh
./task-cli list done
./task-cli list todo
./task-cli list in-progress
```

Valid statuses are `done`, `todo`, and `in-progress`.

## Tests

Run the project tests with:

```sh
go test ./...
```
