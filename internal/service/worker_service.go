package service

import (
	"crud-go/internal/model"
	"crud-go/internal/model/errors"
	"crud-go/internal/repository"
	"fmt"
)

type IWorkerService interface {
	GetAllOrders(string, int) ([]model.Order, error)
	GetOrderById(string, int, int) (model.Order, error)
	SetTaskCompleted(string, int, int) error
}

// WorkerService предназначен для выполнения операций, требующихся для корректной работы функционала работника.
type WorkerService struct {
	OrderRepository *repository.IOrderRepository
	TaskRepository  *repository.ITaskRepository
}

// NewWorkerService создаёт новый сервис для выполнения операций работника.
//
// Принимает интерфейсы репозиториев для хранения заказов и задач,
// возвращает новый экземпляр сервиса.
func NewWorkerService(or repository.IOrderRepository, tr repository.ITaskRepository) WorkerService {
	return WorkerService{OrderRepository: &or, TaskRepository: &tr}
}

// GetAllOrders служит для получения всех заказов, назначенных данному рабочему.
//
// Принимает логин и роль рабочего,
// возвращает список заказов, назначенных данному рабочему и возможную ошибку.
func (ws WorkerService) GetAllOrders(login string, role int) ([]model.Order, error) {
	if role != 2 {
		return nil,
			fmt.Errorf("You are not authorized for this operation")
	}
	or := *ws.OrderRepository
	tr := *ws.TaskRepository

	orders, ordersErr := or.GetOrdersByWorker(login)
	if ordersErr != nil {
		return nil,
			fmt.Errorf("Error getting orders by Worker: \n%w", ordersErr)
	}

	tasks, tasksErr := tr.GetAllTasks()
	if tasksErr != nil {
		return nil,
			fmt.Errorf("Error getting tasks by Worker:\n%w", tasksErr)
	}

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
// Принимает логин рабочего, его роль и идентификатор заказа,
// возвращает информацию о заказе и возможную ошибку.
func (ws WorkerService) GetOrderById(login string, role int, orderId int) (model.Order, error) {
	if role != 2 {
		return model.Order{}, fmt.Errorf("You are not authorized for this operation")
	}
	or := *ws.OrderRepository
	tr := *ws.TaskRepository

	order, ordersErr := or.GetOrderById(orderId)
	if ordersErr != nil {
		return model.Order{},
			fmt.Errorf("Error getting orders by Worker: \n%w", ordersErr)
	}
	if order.WorkerLogin != login {
		return model.Order{},
			errors.AccessRejectedError{}
	}

	tasks, tasksErr := tr.GetTasksByContracts([]int{orderId})
	if tasksErr != nil {
		return model.Order{},
			fmt.Errorf("Error getting tasks by Worker:\n%w", tasksErr)
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

// SetTaskCompleted служит для пометки о выполнении назначенной работнику задачи.
//
// Принимает логин рабочего, его роль и идентификатор задачи,
// возвращает возможную ошибку.
func (ws WorkerService) SetTaskCompleted(login string, role int, taskId int) error {
	if role != 2 {
		return fmt.Errorf("You are not authorized for this operation")
	}
	or := *ws.OrderRepository
	tr := *ws.TaskRepository

	task, taskErr := tr.GetTaskById(taskId)
	if taskErr != nil {
		return fmt.Errorf("Error getting task: \n%w", taskErr)
	}

	order, orderErr := or.GetOrderById(task.OrderID)
	if orderErr != nil {
		return fmt.Errorf("Error getting order: \n%w", orderErr)
	}
	if order.WorkerLogin != login {
		return errors.AccessRejectedError{}
	}

	tasks, tasksErr := tr.GetTasksByContracts([]int{order.ID})
	if tasksErr != nil {
		return fmt.Errorf("Error getting task: \n%w", tasksErr)
	}
	tx, txErr := tr.Begin()
	if txErr != nil {
		return fmt.Errorf("Error starting transaction: \n%w", txErr)
	}
	var taskToChange *model.Task
	for i, tsk := range tasks {
		if tsk.ID == taskId {
			taskToChange = &tasks[i]
			break
		}
	}

	taskToChange.Finished = true
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
	if priceUnfinished == 0 {
		order.Status = 2
	}

	_, orderSaveErr := or.SaveOrder(order)
	if orderSaveErr != nil {
		tx.Rollback()
		return fmt.Errorf("Error updating order:\n%w", orderSaveErr)
	}

	_, taskSaveErr := tr.SaveTask(*taskToChange)
	if taskSaveErr != nil {
		tx.Rollback()
		return fmt.Errorf("Error updating task:\n%w", taskSaveErr)
	}

	tx.Commit()
	return nil
}
