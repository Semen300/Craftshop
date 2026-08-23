package repository

import (
	"context"
	"crud-go/internal/model"
	"database/sql"
	"fmt"
)

type IBasketRepository interface {
	GetBasket(string) ([]model.TaskCreationDTO, error)
	SaveToBasket(string, model.TaskCreationDTO) error
	DeleteFromBasket(string, int) error
	ClearBasket(string) error

	Begin() (Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

type BasketRepository struct {
	Conn *sql.DB
}

// NewBasketRepository создаёт новый репозиторий для доступа к функционалу корзины.
// Также проводит инициализацию таблицы "basket_items".
//
// Принимает указатель на подключение к базе данных,
// возвращает новый экземпляр репозитория и возможную ошибку.
func NewBasketRepository(db *sql.DB) (BasketRepository, error) {
	basketCreationQuery := `CREATE TABLE IF NOT EXISTS basket_items(
	customerLogin TEXT,
	itemID SERIAL,
	name TEXT,
	itemPrice INTEGER,
	amount INTEGER,
	PRIMARY KEY (customerLogin, itemID))`

	_, creationErr := db.Exec(basketCreationQuery)
	if creationErr != nil {
		return BasketRepository{}, fmt.Errorf("Error creating table 'basket_items': \n%w", creationErr)
	}
	return BasketRepository{db}, nil
}

// Begin служит для начала транзакции.
//
// Не принимает значений,
// возвращает новую транзакцию и возможную ошибку.
func (br BasketRepository) Begin() (Tx, error) {
	return br.Conn.Begin()
}

// BeginTx служит для начала транзакции с указанным контекстом и параметрами.
//
// Принимает контекст и параметры транзакции,
// возвращает новую транзакцию и возможную ошибку.
func (br BasketRepository) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	return br.Conn.BeginTx(ctx, opts)
}

// GetBasket служит для получения всех товаров из корзины покупателя.
//
// Принимает логин покупателя,
// возвращает список товаров в корзине и возможную ошибку.
func (br BasketRepository) GetBasket(customerLogin string) ([]model.TaskCreationDTO, error) {
	query := `SELECT bi.itemID, bi.name, bi.itemPrice, bi.amount
	FROM basket_items bi
	WHERE customerLogin = $1`

	rows, queryErr := br.Conn.Query(query, customerLogin)
	if queryErr != nil {
		return nil, fmt.Errorf("Error executing query \"%s\" to table \"basket_items\":\n %w", query, queryErr)
	}
	tasks := make([]model.TaskCreationDTO, 0)
	for rows.Next() {
		var taskDTO model.TaskCreationDTO
		scanErr := rows.Scan(&taskDTO.ItemID, &taskDTO.Name, &taskDTO.ItemPrice, &taskDTO.Amount)
		if scanErr != nil {
			return nil, fmt.Errorf("Error scanning values from rows: \n%w", scanErr)
		}
		tasks = append(tasks, taskDTO)
	}

	if iterErr := rows.Err(); iterErr != nil {
		return nil, fmt.Errorf("Error processing rows: \n%w", iterErr)
	}

	return tasks, nil
}

// SaveToBasket служит для добавления товара в корзину покупателя или обновления его данных.
//
// Принимает логин покупателя и данные сохраняемого товара,
// возвращает возможную ошибку.
func (br BasketRepository) SaveToBasket(customerLogin string, toSave model.TaskCreationDTO) error {
	selectQuery := `SELECT *
	FROM basket_items
	WHERE customerLogin = $1 AND itemID = $2`
	rows, selectQueryErr := br.Conn.Query(selectQuery, customerLogin, toSave.ItemID)
	if selectQueryErr != nil {
		return fmt.Errorf("Error executing query \"%s\" to table \"basket_items\":\n %w", selectQuery, selectQueryErr)
	}
	isPresent := rows.Next()
	saveQuery := ""
	if isPresent {
		saveQuery = `UPDATE basket_items
		SET name = $3,
		itemPrice = $4,
		amount = $5 
		WHERE customerLogin = $1 AND itemID = $2`
	} else {
		saveQuery = `INSERT INTO basket_items 
		VALUES ($1, $2, $3, $4, $5)`
	}

	_, queryErr := br.Conn.Exec(saveQuery,
		customerLogin,
		toSave.ItemID,
		toSave.Name,
		toSave.ItemPrice,
		toSave.Amount)

	if queryErr != nil {
		return fmt.Errorf("Error executing query \"%s\" to table \"basket_items\":\n %w", saveQuery, queryErr)
	}

	return nil
}

// DeleteFromBasket служит для удаления товара из корзины покупателя.
//
// Принимает логин покупателя и ID удаляемого товара,
// возвращает возможную ошибку.
func (br BasketRepository) DeleteFromBasket(customerLogin string, itemID int) error {
	query := `DELETE FROM basket_items
	WHERE customerLogin = $1 AND itemID = $2`
	_, queryErr := br.Conn.Exec(query, customerLogin, itemID)
	if queryErr != nil {
		return fmt.Errorf("Error executing query \"%s\" to table \"basket_items\":\n %w", query, queryErr)
	}
	return nil
}

// ClearBasket служит для удаления всех товаров из корзины покупателя.
//
// Принимает логин покупателя,
// возвращает возможную ошибку.
func (br BasketRepository) ClearBasket(customerLogin string) error {
	query := `DELETE FROM basket_items
	WHERE customerLogin = $1`
	_, queryErr := br.Conn.Exec(query, customerLogin)
	if queryErr != nil {
		return fmt.Errorf("Error executing query \"%s\" to table \"basket_items\":\n %w", query, queryErr)
	}
	return nil
}
