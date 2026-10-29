package service

import (
	"crud-go/internal/model"
	"crud-go/internal/model/errors"
	"crud-go/internal/repository"
	"fmt"
)

type IManagerService interface {
	GetAllWorkers(string, int) ([]model.Worker, error)
	GetAllOrders(string, int) ([]model.Order, error)
	GetOrderById(string, int, int) (model.Order, error)
	AssignWorkerToOrder(string, int, int, string) error
}

type ManagerService struct {
	OrderRepository *repository.IOrderRepository
	TaskRepository  *repository.ITaskRepository
	UserRepository  *repository.IUserRepository
}

// NewManagerService создаёт новый сервис для выполнения операций менеджера.
//
// Принимает интерфейсы репозиториев для хранения заказов, задач и пользователей,
// возвращает новый экземпляр сервиса.
func NewManagerService(or repository.IOrderRepository, tr repository.ITaskRepository, ur repository.IUserRepository) ManagerService {
	return ManagerService{OrderRepository: &or, TaskRepository: &tr, UserRepository: &ur}
}

// GetAllWorkers служит для получения всех рабочих, подчинённых данному менеджеру.
//
// Принимает логин и роль менеджера,
// возвращает список рабочих и возможную ошибку.
func (ms ManagerService) GetAllWorkers(login string, role int) ([]model.Worker, error) {
	if role != 3 {
		return nil,
			fmt.Errorf("You are not authorized for this operation")
	}

	ur := *ms.UserRepository

	workers, getErr := ur.GetWorkersByManager(login)
	if getErr != nil {
		return nil,
			fmt.Errorf("Error getting workers by manager: \n%w", getErr)
	}

	return workers, nil
}

// GetAllOrders служит для получения всех заказов, доступных данному менеджеру.
//
// Принимает логин и роль менеджера,
// возвращает список заказов и возможную ошибку.
func (ms ManagerService) GetAllOrders(login string, role int) ([]model.Order, error) {
	if role != 3 {
		return nil, fmt.Errorf("You are not authorized for this operation")
	}

	or := *ms.OrderRepository
	tr := *ms.TaskRepository

	managerOrders, ordersErr := or.GetOrdersByManager(login)
	emptyOrders, ordersErr := or.GetOrdersByManager("")
	if ordersErr != nil {
		return nil,
			fmt.Errorf("Error getting orders for manager %s:\n%w", login, ordersErr)
	}

	var ordersID []int
	for _, order := range managerOrders {
		ordersID = append(ordersID, order.ID)
	}
	for _, order := range emptyOrders {
		ordersID = append(ordersID, order.ID)
	}

	tasks, tasksErr := tr.GetTasksByContracts(ordersID)
	if tasksErr != nil {
		return nil,
			fmt.Errorf("Error getting tasks:\n%w", tasksErr)
	}

	orders := append(managerOrders, emptyOrders...)
	tasksMap := make(map[int][]model.Task, 0)
	for _, task := range tasks {
		tasksMap[task.OrderID] = append(tasksMap[task.OrderID], task)
	}

	for i, order := range orders {
		orders[i].Tasks = tasksMap[order.ID]
		var numberItemsCompleted, numberItemsAll int
		var priceUnfinished int
		for _, task := range tasksMap[order.ID] {
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
// Принимает логин менеджера, его роль и идентификатор заказа,
// возвращает информацию о заказе и возможную ошибку.
func (ms ManagerService) GetOrderById(login string, role int, id int) (model.Order, error) {
	if role != 3 {
		return model.Order{}, fmt.Errorf("You are not authorized for this operation")
	}

	or := *ms.OrderRepository
	tr := *ms.TaskRepository

	order, orderErr := or.GetOrderById(id)
	if orderErr != nil {
		return model.Order{},
			fmt.Errorf("Error getting order by ID: \n%w", orderErr)
	}
	if order.ManagerLogin != login && order.ManagerLogin != "" {
		return model.Order{},
			errors.AccessRejectedError{}
	}

	tasks, tasksErr := tr.GetTasksByContracts([]int{id})
	if tasksErr != nil {
		return model.Order{},
			fmt.Errorf("Error getting tasks for order: \n%w", tasksErr)
	}
	order.Tasks = tasks
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

	return order, nil
}

// AssignWorkerToOrder служит для назначения рабочего на заказ.
//
// Принимает логин менеджера, его роль, идентификатор заказа и логин рабочего,
// возвращает возможную ошибку.
func (ms ManagerService) AssignWorkerToOrder(login string, role int, orderId int, workerLogin string) error {
	if role != 3 {
		return fmt.Errorf("You are not authorized for this operation")
	}

	or := *ms.OrderRepository
	ur := *ms.UserRepository

	order, orderGetErr := or.GetOrderById(orderId)
	if orderGetErr != nil {
		return fmt.Errorf("Error getting order by ID: \n%w", orderGetErr)
	}
	if order.ManagerLogin != login && order.ManagerLogin != "" {
		return errors.AccessRejectedError{}
	}

	worker, workerGetErr := ur.GetWorker(workerLogin)
	if workerGetErr != nil {
		return fmt.Errorf("Error getting user by login: \n%w", orderGetErr)
	}
	if worker.SuperiorLogin != login {
		return errors.AccessRejectedError{}
	}

	order.ManagerLogin = login
	order.WorkerLogin = workerLogin
	order.Status = 1
	_, saveErr := or.SaveOrder(order)
	if saveErr != nil {
		return fmt.Errorf("Error saving order: \n%w", saveErr)
	}
	return nil
}
