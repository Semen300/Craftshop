package service

import (
	"crud-go/internal/model"
	"crud-go/internal/model/errors"
	"crud-go/internal/repository"
	"fmt"
)

type ICustomerService interface {
	GetOrdersByCustomer(login string, role int) ([]model.Order, error)
	GetOrderById(login string, role int, id int) (model.Order, error)
	CreateOrder(login string, role int, orderDTO model.OrderCreationDTO) (int, error)
	DeleteOrder(login string, role int, id int) (int, error)
	GetItems(login string, role int) ([]model.Item, error)
	GetBasket(login string, role int) ([]model.TaskCreationDTO, error)
	SaveToBasket(login string, role int, item model.TaskCreationDTO) error
	DeleteFromBasket(login string, role int, id int) error
	ClearBasket(login string, role int) error
}

type CustomerService struct {
	OrderRepository  *repository.IOrderRepository
	TaskRepository   *repository.ITaskRepository
	BasketRepository *repository.IBasketRepository
	ItemRepository   *repository.IItemRepository
}

// NewCustomerService создаёт новый сервис для выполнения операций покупателя.
//
// Принимает интерфейсы репозиториев для хранения заказов, задач, корзины и товаров,
// возвращает новый экземпляр сервиса.
func NewCustomerService(or repository.IOrderRepository, tr repository.ITaskRepository, br repository.IBasketRepository, ir repository.IItemRepository) CustomerService {
	return CustomerService{OrderRepository: &or, TaskRepository: &tr, BasketRepository: &br, ItemRepository: &ir}
}

// GetOrdersByCustomer служит для получения всех заказов покупателя.
//
// Принимает логин и роль покупателя,
// возвращает список заказов покупателя и возможную ошибку.
func (cs CustomerService) GetOrdersByCustomer(login string, role int) ([]model.Order, error) {
	if role != 1 {
		return nil, fmt.Errorf("You are not authorized for this operation")
	}
	or := *(cs.OrderRepository)
	tr := *(cs.TaskRepository)

	orders, orderRepoErr := or.GetOrdersByCustomer(login)
	if orderRepoErr != nil {
		return nil, fmt.Errorf("Error getting orders by customer login: \n%w", orderRepoErr)
	}
	var ordersID []int
	for _, order := range orders {
		ordersID = append(ordersID, order.ID)
	}

	tasks, taskRepoErr := tr.GetTasksByContracts(ordersID)
	if taskRepoErr != nil {
		return nil, fmt.Errorf("Error getting tasks from TaskRepository: \n%w", taskRepoErr)
	}
	tasksByOrder := make(map[int][]model.Task)
	for _, task := range tasks {
		tasksByOrder[task.OrderID] = append(tasksByOrder[task.OrderID], task)
	}

	for i, order := range orders {
		orders[i].Tasks = tasksByOrder[order.ID]
		var numberItemsCompleted, numberItemsAll int
		var priceUnfinished int
		for _, task := range tasksByOrder[order.ID] {
			if task.Finished {
				numberItemsCompleted += task.Amount
			} else {
				priceUnfinished += task.Price
			}
			numberItemsAll += task.Amount
		}
		orders[i].PercentOfComplition = float64(numberItemsCompleted) / float64(numberItemsAll)
		orders[i].PriceUnfinished = priceUnfinished
	}
	return orders, nil
}

// GetOrderById служит для получения заказа по его идентификатору.
//
// Принимает логин покупателя, его роль и идентификатор заказа,
// возвращает информацию о заказе и возможную ошибку.
func (cs CustomerService) GetOrderById(login string, role int, orderId int) (model.Order, error) {
	if role != 1 {
		return model.Order{}, fmt.Errorf("You are not authorized for this operation")
	}
	or := *(cs.OrderRepository)
	tr := *(cs.TaskRepository)

	order, orderRepoErr := or.GetOrderById(orderId)
	if orderRepoErr != nil {
		return model.Order{}, fmt.Errorf("Error getting order by order ID: \n%w", orderRepoErr)
	}
	if order.CustomerLogin != login {
		return model.Order{},
			errors.AccessRejectedError{}
	}

	tasks, taskRepoErr := tr.GetTasksByContracts([]int{orderId})
	if taskRepoErr != nil {
		return model.Order{}, fmt.Errorf("Error getting tasks by order ID: \n%w", taskRepoErr)
	}
	var numberItemsCompleted, numberItemsAll int
	var priceUnfinished int
	for _, task := range tasks {
		if task.Finished {
			numberItemsCompleted += task.Amount
		} else {
			priceUnfinished += task.Price
		}
		numberItemsAll += task.Amount
	}

	order.PercentOfComplition = float64(numberItemsCompleted) / float64(numberItemsAll)
	order.PriceUnfinished = priceUnfinished
	order.Tasks = tasks
	return order, nil
}

// CreateOrder служит для создания нового заказа покупателя.
//
// Принимает логин и роль покупателя, а также данные создаваемого заказа,
// возвращает идентификатор созданного заказа и возможную ошибку.
func (cs CustomerService) CreateOrder(login string, role int, orderDTO model.OrderCreationDTO) (int, error) {
	if role != 1 {
		return 0, fmt.Errorf("You are not authorized for this operation")
	}
	or := *(cs.OrderRepository)
	tr := *(cs.TaskRepository)

	price := 0

	for _, item := range orderDTO.Tasks {
		price += item.ItemPrice * item.Amount
	}
	genericName := fmt.Sprintf("Заказ для пользователя %s до %02d.%02d.%4d на сумму %d.%d", login, orderDTO.Deadline.Day(), orderDTO.Deadline.Month(), orderDTO.Deadline.Year(), price/100, price%100)
	orderToSave := model.Order{
		Name:          genericName,
		Deadline:      orderDTO.Deadline,
		CustomerLogin: login,
		PriceTotal:    price,
		Status:        0,
	}

	tx, txErr := tr.Begin()
	if txErr != nil {
		return 0,
			fmt.Errorf("Error starting transaction: %w", txErr)
	}
	createdId, orderRepoErr := or.SaveOrder(orderToSave)
	if orderRepoErr != nil {
		tx.Rollback()
		return 0, fmt.Errorf("Error during saving order: \n%w", orderRepoErr)
	}

	for _, task := range orderDTO.Tasks {
		taskGenericName := fmt.Sprintf("Задача: изготовление %s до %02d.%02d.%4d", task.Name, orderDTO.Deadline.Day(), orderDTO.Deadline.Month(), orderDTO.Deadline.Year())
		taskToSave := model.Task{
			Name:     taskGenericName,
			OrderID:  createdId,
			ItemID:   task.ItemID,
			Amount:   task.Amount,
			Finished: false,
			Price:    task.Amount * task.ItemPrice,
		}
		_, taskRepoErr := tr.SaveTask(taskToSave)
		if taskRepoErr != nil {
			tx.Rollback()
			return 0, fmt.Errorf("Error during saving task %s: \n%w", taskToSave.ToString(), orderRepoErr)
		}
	}

	tx.Commit()
	return createdId, nil
}

// DeleteOrder служит для удаления заказа покупателя.
//
// Принимает логин и роль покупателя, а также идентификатор заказа,
// возвращает стоимость незавершённых задач и возможную ошибку.
func (cs CustomerService) DeleteOrder(login string, role int, orderId int) (int, error) {
	if role != 1 {
		return 0, fmt.Errorf("You are not authorized for this operation")
	}
	or := *(cs.OrderRepository)
	tr := *(cs.TaskRepository)

	order, getErr := or.GetOrderById(orderId)
	if getErr != nil {
		return 0, getErr
	}
	if order.CustomerLogin != login {
		return 0, errors.AccessRejectedError{}
	}

	var priceUnfinished int
	tasks, taskRepoErr := tr.GetTasksByContracts([]int{orderId})
	if taskRepoErr != nil {
		return 0, fmt.Errorf("Error getting tasks by order ID: \n%w", taskRepoErr)
	}

	tx, txErr := tr.Begin()
	if txErr != nil {
		return 0, fmt.Errorf("Error starting transaction")
	}
	for _, task := range tasks {
		if !task.Finished {
			priceUnfinished += task.Price
		}
		taskDeleteErr := tr.DeleteTask(task.ID)
		if taskDeleteErr != nil {
			tx.Rollback()
			return 0, fmt.Errorf("Error deleting task %s by order ID: \n%w", task.ToString(), taskDeleteErr)
		}
	}
	orderDeleteErr := or.DeleteOrder(orderId)
	if orderDeleteErr != nil {
		tx.Rollback()
		return 0, fmt.Errorf("Error deleting order by ID: \n%w", orderDeleteErr)
	}
	tx.Commit()
	return priceUnfinished, nil
}

// GetItems служит для получения списка товаров.
//
// Принимает логин и роль покупателя,
// возвращает список товаров и возможную ошибку.
func (cs CustomerService) GetItems(login string, role int) ([]model.Item, error) {
	if role != 1 {
		return nil, fmt.Errorf("You are not authorized for this operation")
	}
	ir := *(cs.ItemRepository)

	items, itemRepoErr := ir.GetAllItems()
	if itemRepoErr != nil {
		return nil, fmt.Errorf("Error getting items: \n%w", itemRepoErr)
	}

	return items, nil
}

// GetBasket служит для получения корзины покупателя.
//
// Принимает логин и роль покупателя,
// возвращает список товаров в корзине и возможную ошибку.
func (cs CustomerService) GetBasket(login string, role int) ([]model.TaskCreationDTO, error) {
	if role != 1 {
		return nil,
			fmt.Errorf("You are not authorized for this operation")
	}
	br := *(cs.BasketRepository)

	basket, basketErr := br.GetBasket(login)
	if basketErr != nil {
		return nil,
			fmt.Errorf("Error getting customer basket: \n%w", basketErr)
	}

	return basket, nil
}

// SaveToBasket служит для добавления товара в корзину покупателя.
//
// Принимает логин и роль покупателя, а также данные товара,
// возвращает возможную ошибку.
func (cs CustomerService) SaveToBasket(login string, role int, item model.TaskCreationDTO) error {
	if role != 1 {
		return fmt.Errorf("You are not authorized for this operation")
	}
	br := *(cs.BasketRepository)

	saveErr := br.SaveToBasket(login, item)
	if saveErr != nil {
		return fmt.Errorf("Error saving item to basket: \n%w", saveErr)
	}
	return nil
}

// DeleteFromBasket служит для удаления товара из корзины покупателя.
//
// Принимает логин и роль покупателя, а также идентификатор товара,
// возвращает возможную ошибку.
func (cs CustomerService) DeleteFromBasket(login string, role int, itemId int) error {
	if role != 1 {
		return fmt.Errorf("You are not authorized for this operation")
	}
	br := *(cs.BasketRepository)

	deleteErr := br.DeleteFromBasket(login, itemId)
	if deleteErr != nil {
		return fmt.Errorf("Error deleting item from basket: \n%w", deleteErr)
	}

	return nil
}

// ClearBasket служит для удаления всех товаров из корзины покупателя.
//
// Принимает логин и роль покупателя,
// возвращает возможную ошибку.
func (cs CustomerService) ClearBasket(login string, role int) error {
	if role != 1 {
		return fmt.Errorf("You are not authorized for this operation")
	}
	br := *cs.BasketRepository
	return br.ClearBasket(login)
}
