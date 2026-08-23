package repository_test

import (
	"crud-go/internal/model"
	"crud-go/internal/repository"
	"database/sql"
	"fmt"
	"log"
	"testing"
)

var expectedTasks = []model.Task{
	{ID: 1, Name: "task1", OrderID: 1, ItemID: 1, Amount: 1, Finished: true, Price: 100},
	{ID: 2, Name: "task2", OrderID: 2, ItemID: 2, Amount: 2, Finished: true, Price: 200},
	{ID: 3, Name: "task3", OrderID: 3, ItemID: 3, Amount: 3, Finished: true, Price: 300},
}

func migrateTasks(db *sql.DB) {
	_, createErr := db.Exec(`CREATE TABLE IF NOT EXISTS tasks(
	id SERIAL PRIMARY KEY,
	name TEXT,
	contractID SERIAL,
	itemID SERIAL,
	amount SERIAL,
	finished BOOL,
	price SERIAL)`)
	if createErr != nil {
		log.Fatal(fmt.Errorf("Error executing migration into table 'tasks': \nError creating table: \n%w", createErr))
	}

	_, insertErr := db.Exec(`INSERT INTO tasks
	VALUES ($1, $2, $3, $4, $5, $6, $7),
	($8, $9, $10, $11, $12, $13, $14),
	($15, $16, $17, $18, $19, $20, $21)`,
		1, "task1", 1, 1, 1, true, 100,
		2, "task2", 2, 2, 2, true, 200,
		3, "task3", 3, 3, 3, true, 300)
	if insertErr != nil {
		log.Fatal(fmt.Errorf("Error executing migration into table 'tasks': \nError adding values: \n%w", insertErr))
	}
}

func resetTasks(repo *repository.TaskRepository) {
	repo.Conn.Exec(`DROP TABLE IF EXISTS tasks`)
	migrateTasks(repo.Conn)
	repo.CurrerntID = 3
}

func getNumberOfTasks(db *sql.DB) int {
	rows, _ := db.Query("SELECT * FROM tasks")
	var numberOfTasks = 0
	for rows.Next() {
		numberOfTasks++
	}
	return numberOfTasks
}

func TestNewTaskRepository(t *testing.T) {
	repo, repoErr := repository.NewTaskRepository(testDB)
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	pingErr := repo.Conn.Ping()
	if pingErr != nil {
		t.Fatal(pingErr)
	}
}

func TestGetAllTasks(t *testing.T) {
	repo, repoErr := repository.NewTaskRepository(testDB)
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	resetTasks(&repo)
	tasks, getErr := repo.GetAllTasks()
	if getErr != nil {
		t.Fatal(getErr)
	}
	if len(tasks) != len(expectedTasks) {
		t.Errorf("Result length missmatch: expected %d, got %d", len(expectedTasks), len(tasks))
	}
	for i, task := range tasks {
		if task != expectedTasks[i] {
			t.Errorf("Result missmatch: expected %s, got %s", expectedTasks[i].ToString(), task.ToString())
		}
	}
}

func TestGetTasksByContract(t *testing.T) {
	var rowNum = 1
	repo, repoErr := repository.NewTaskRepository(testDB)
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	resetTasks(&repo)
	tasks, getErr := repo.GetTasksByContracts([]int{rowNum})
	if getErr != nil {
		t.Fatal(getErr)
	}
	if len(tasks) != 1 {
		t.Errorf("Result length missmatch: expected %d, got %d", 1, len(tasks))
	}
	for i, task := range tasks {
		if task != expectedTasks[i] {
			t.Errorf("Result missmatch: expected %s, got %s", expectedTasks[i].ToString(), task.ToString())
		}
	}
}

func TestGetTaskById(t *testing.T) {
	var rowNum = 2
	repo, repoErr := repository.NewTaskRepository(testDB)
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	resetTasks(&repo)
	task, getErr := repo.GetTaskById(rowNum)
	if getErr != nil {
		t.Fatal(getErr)
	}
	if task != expectedTasks[rowNum-1] {
		t.Errorf("Result missmatch: expected %s, got %s", expectedTasks[rowNum-1].ToString(), task.ToString())
	}
}

func TestSaveTaskSave(t *testing.T) {
	var taskToSave = model.Task{ID: 4, Name: "task4", OrderID: 4, ItemID: 4, Amount: 4, Finished: true, Price: 400}
	repo, repoErr := repository.NewTaskRepository(testDB)
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	resetTasks(&repo)
	_, saveErr := repo.SaveTask(taskToSave)
	if saveErr != nil {
		t.Fatal(saveErr)
	}

	numberOfTasks := getNumberOfTasks(repo.Conn)
	if numberOfTasks != 4 {
		t.Errorf("Result length missmatch: expected %d, got %d", 4, numberOfTasks)
	}

	task, _ := repo.GetTaskById(4)
	if task != taskToSave {
		t.Errorf("Result missmatch: expected %s, got %s", taskToSave.ToString(), task.ToString())
	}
}

func TestSaveTaskUpdate(t *testing.T) {
	var taskToSave = model.Task{ID: 3, Name: "task4", OrderID: 4, ItemID: 4, Amount: 4, Finished: true, Price: 400}
	repo, repoErr := repository.NewTaskRepository(testDB)
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	resetTasks(&repo)
	_, saveErr := repo.SaveTask(taskToSave)
	if saveErr != nil {
		t.Fatal(saveErr)
	}

	numberOfTasks := getNumberOfTasks(repo.Conn)
	if numberOfTasks != 3 {
		t.Errorf("Result length missmatch: expected %d, got %d", 4, numberOfTasks)
	}

	task, _ := repo.GetTaskById(3)
	if task != taskToSave {
		t.Errorf("Result missmatch: expected %s, got %s", taskToSave.ToString(), task.ToString())
	}
}

func TestDeleteTask(t *testing.T) {
	var rowNum = 2
	repo, repoErr := repository.NewTaskRepository(testDB)
	if repoErr != nil {
		t.Fatal(repoErr)
	}
	resetTasks(&repo)
	deleteErr := repo.DeleteTask(rowNum)
	if deleteErr != nil {
		t.Fatal(deleteErr)
	}

	task, _ := repo.GetTaskById(rowNum)
	if task != (model.Task{}) {
		t.Error("Delete error: row still exists")
	}
}
