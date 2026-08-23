package repository

import (
	"crud-go/internal/model"
	"database/sql"
	"fmt"
)

type IUserRepository interface {
	GetWorkersByManager(managerLogin string) (workerInfo []model.Worker, e error)
	GetRoleByLogin(userLogin string) (role int, hashedPass string, e error)
	GetWorker(workerLogin string) (workerInfo model.Worker, e error)
	SaveCustomer(customerToSave model.Customer) error
}

// UserRepository предназначен для выполнения операций, требующих доступа к БД, хранящей данные о пользователях.
type UserRepository struct {
	Conn *sql.DB
}

// NewUserRepository создаёт новый репозиторий для доступа к функционалу пользователей.
// Также проводит инициализацию таблиц "managers", "workers" и "customers".
//
// Принимает указатель на подключение к базе данных,
// возвращает новый экземпляр репозитория и возможную ошибку.
func NewUserRepository(db *sql.DB) (UserRepository, error) {
	managerQuery := `CREATE TABLE IF NOT EXISTS managers(
	login TEXT PRIMARY KEY,
	password TEXT,
	fio TEXT
	)`
	workerQuery := `CREATE TABLE IF NOT EXISTS workers (
	login TEXT PRIMARY KEY,
	password TEXT,
	fio TEXT,
	superiorLogin TEXT
	)`
	customerQuery := `CREATE TABLE IF NOT EXISTS customers (
	login TEXT PRIMARY KEY,
	password TEXT,
	fio TEXT,
	number TEXT,
	email TEXT
	)`

	_, managerErr := db.Exec(managerQuery)
	if managerErr != nil {
		return UserRepository{},
			fmt.Errorf("Error creating table \"managers\":\n %w", managerErr)
	}

	_, workerErr := db.Exec(workerQuery)
	if workerErr != nil {
		return UserRepository{},
			fmt.Errorf("Error creating table \"workers\":\n %w", workerErr)
	}

	_, customerErr := db.Exec(customerQuery)
	if customerErr != nil {
		return UserRepository{},
			fmt.Errorf("Error creating table \"customers\":\n %w", customerErr)
	}

	return UserRepository{db}, nil
}

// GetWorkersByManager служит для получения всех рабочих, назначенных определённому менеджеру.
//
// Принимает логин менеджера, для которого нужно получить работников,
// возвращает список работников и возможную ошибку.
func (u UserRepository) GetWorkersByManager(managerLogin string) ([]model.Worker, error) {
	query := `SELECT *
	FROM workers
	WHERE superiorLogin = $1`
	rows, queryErr := u.Conn.Query(query, managerLogin)
	if queryErr != nil {
		return nil,
			fmt.Errorf("Error executing query \"%s\" to table \"workers\":\n %w", query, queryErr)
	}
	defer rows.Close()

	var workers = make([]model.Worker, 0)
	for rows.Next() {
		var worker model.Worker
		scanErr := rows.Scan(&worker.Login, &worker.Password, &worker.Fio, &worker.SuperiorLogin)
		if scanErr != nil {
			return nil,
				fmt.Errorf("Error scanning values from rows:\n %w", scanErr)
		}
		workers = append(workers, worker)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil,
			fmt.Errorf("Error processing rows: \n %w", iterErr)
	}
	return workers, nil
}

// GetRoleByLogin служит для получения информации о пользователе (его роль и пароль) по его логину.
//
// Принимает логин пользователя, по которому нужно получить информацию,
// возвращает роль пользователя, его пароль и возможную ошибку.
func (u UserRepository) GetRoleByLogin(login string) (int, string, error) {
	managerQuery := `SELECT password FROM managers WHERE login = $1`
	workerQuery := `SELECT password FROM workers WHERE login = $1`
	customerQuery := `SELECT password FROM customers WHERE login = $1`

	managerRows, managerErr := u.Conn.Query(managerQuery, login)
	if managerErr != nil {
		return 0, "",
			fmt.Errorf("Error executing query \"%s\" to table \"managers\":\n %w", managerQuery, managerErr)
	}
	defer managerRows.Close()

	workerRows, workerErr := u.Conn.Query(workerQuery, login)
	if workerErr != nil {
		return 0, "",
			fmt.Errorf("Error executing query \"%s\" to table \"workers\":\n %w", workerQuery, workerErr)
	}
	defer workerRows.Close()

	customerRows, customerErr := u.Conn.Query(customerQuery, login)
	if customerErr != nil {
		return 0, "",
			fmt.Errorf("Error executing query \"%s\" to table \"customers\":\n %w", customerQuery, customerErr)
	}
	defer customerRows.Close()

	role := 0
	password := ""

	if manRowsErr := managerRows.Err(); managerRows.Next() && manRowsErr == nil {
		role = 3
		managerRows.Scan(&password)
	} else if worRowsErr := workerRows.Err(); workerRows.Next() && worRowsErr == nil {
		role = 2
		workerRows.Scan(&password)
	} else if cusRowsErr := customerRows.Err(); customerRows.Next() && cusRowsErr == nil {
		role = 1
		customerRows.Scan(&password)
	} else if manRowsErr != nil || worRowsErr != nil || cusRowsErr != nil {
		return 0, "",
			fmt.Errorf("Error processing rows: \n")
	}

	return role, password, nil
}

// GetWorker служит для получения информации о работнике.
//
// Принимает логин рабочего, о котором нужно получить информацию,
// возвращает информацию о рабочем и возможную ошибку
func (u UserRepository) GetWorker(workerLogin string) (model.Worker, error) {
	query := `SELECT * 
	FROM workers
	WHERE id = $1`

	var worker model.Worker
	queryErr := u.Conn.QueryRow(query, workerLogin).Scan(&worker.Login, &worker.Password, &worker.Fio, &worker.SuperiorLogin)
	if queryErr != nil {
		return model.Worker{}, queryErr
	}
	return worker, nil
}

// SaveCustomer служит для идемпотентного сохранения покупателя.
// В случае, если сохраняемый покупатель существует в БД, обновляет его,
// иначе создаёт нового покупателя.
//
// Принимает покупателя, которого нужно сохранить,
// возвращает возможную ошибку.
func (u UserRepository) SaveCustomer(customer model.Customer) error {
	selectQuery := `SELECT login 
	FROM customers
	WHERE login = $1`

	rows, queryErr := u.Conn.Query(selectQuery, customer.Login)
	if queryErr != nil {
		return fmt.Errorf("Error executing query \"%s\" to table \"orders\":\n %w", selectQuery, queryErr)
	}
	defer rows.Close()

	isPresent := rows.Next()

	var saveQuery string
	if !isPresent {
		saveQuery = `INSERT INTO customers
		VALUES ($1, $2, $3, $4, $5)`
	} else {
		saveQuery = `UPDATE customers
		SET password = $2,
		fio = $3,
		number = $4,
		email = $5
		WHERE login = $1`
	}

	_, saveErr := u.Conn.Exec(saveQuery,
		customer.Login,
		customer.Password,
		customer.Fio,
		customer.Number,
		customer.Email)

	if saveErr != nil {
		return fmt.Errorf("Error executing query \"%s\" to table \"orders\":\n %w", saveQuery, saveErr)
	}

	return nil
}
