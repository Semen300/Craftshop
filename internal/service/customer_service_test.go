package service_test

import (
	"crud-go/internal/model"
	"crud-go/internal/service"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCustomerService(t *testing.T) {
	orderMock := new(orderRepoMock)
	taskMock := new(taskRepoMock)
	basketMock := new(basketRepoMock)
	itemMock := new(itemRepoMock)
	testService := service.NewCustomerService(orderMock, taskMock, basketMock, itemMock)

	assert.NotEqual(t, service.CustomerService{}, testService)
}

func TestGetOrderByCustomer(t *testing.T) {
	orderRepoMock := new(orderRepoMock)
	taskRepoMock := new(taskRepoMock)
	testService := service.NewCustomerService(orderRepoMock, taskRepoMock, nil, nil)

	mockOrders := []model.Order{
		{
			ID:            1,
			Name:          "order1",
			Deadline:      time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
			CustomerLogin: "customer1",
			PriceTotal:    10000,
			Status:        1,
		},
		{
			ID:            2,
			Name:          "order2",
			Deadline:      time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
			ManagerLogin:  "manager1",
			WorkerLogin:   "worker1",
			CustomerLogin: "customer1",
			PriceTotal:    20000,
			Status:        2,
		},
	}

	mockTasks := []model.Task{
		{ID: 1, Name: "task1", OrderID: 1, ItemID: 1, Amount: 10, Finished: false, Price: 5000},
		{ID: 2, Name: "task2", OrderID: 1, ItemID: 2, Amount: 20, Finished: true, Price: 5000},
		{ID: 3, Name: "task3", OrderID: 2, ItemID: 1, Amount: 10, Finished: false, Price: 5000},
		{ID: 4, Name: "task4", OrderID: 2, ItemID: 2, Amount: 60, Finished: true, Price: 15000},
	}

	orderRepoMock.On("GetOrdersByCustomer", "customer1").Return(mockOrders, nil)
	taskRepoMock.On("GetTasksByOrders", []int{1, 2}).Return(mockTasks, nil)

	gotOrders, ordersErr := testService.GetOrdersByCustomer("customer1", 1)
	if ordersErr != nil {
		t.Fatal(ordersErr)
	}

	expectedOrders := []model.Order{
		{
			ID:                  1,
			Name:                "order1",
			Deadline:            time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
			CustomerLogin:       "customer1",
			PriceTotal:          10000,
			Status:              1,
			PercentOfComplition: float64(2) / 3,
			PriceUnfinished:     5000,
			Tasks: []model.Task{
				mockTasks[0],
				mockTasks[1],
			},
		},
		{
			ID:                  2,
			Name:                "order2",
			Deadline:            time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
			ManagerLogin:        "manager1",
			WorkerLogin:         "worker1",
			CustomerLogin:       "customer1",
			PriceTotal:          20000,
			Status:              2,
			PercentOfComplition: float64(6) / 7,
			PriceUnfinished:     5000,
			Tasks: []model.Task{
				mockTasks[2],
				mockTasks[3],
			},
		},
	}

	assert.Equal(t, 2, len(gotOrders))
	assert.Equal(t, expectedOrders, gotOrders)
}

func TestGetOrderById(t *testing.T) {
	orderRepoMock := new(orderRepoMock)
	taskRepoMock := new(taskRepoMock)
	testService := service.NewCustomerService(orderRepoMock, taskRepoMock, nil, nil)

	mockOrder := model.Order{
		ID:            1,
		Name:          "order1",
		Deadline:      time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
		ManagerLogin:  "manager1",
		WorkerLogin:   "worker1",
		CustomerLogin: "customer1",
		PriceTotal:    10000,
		Status:        2,
	}
	mockTasks := []model.Task{
		{
			ID:       1,
			Name:     "task1",
			OrderID:  1,
			ItemID:   1,
			Amount:   10,
			Finished: true,
			Price:    5000,
		},
		{
			ID:       2,
			Name:     "task2",
			OrderID:  1,
			ItemID:   2,
			Amount:   20,
			Finished: false,
			Price:    5000,
		},
	}

	orderRepoMock.On("GetOrderById", 1).Return(mockOrder, nil)
	taskRepoMock.On("GetTasksByContracts", []int{1}).Return(mockTasks, nil)

	gotOrder, orderErr := testService.GetOrderById("customer1", 1, 1)
	if orderErr != nil {
		t.Fatal(orderErr)
	}
	expectedOrder := model.Order{
		ID:                  1,
		Name:                "order1",
		Deadline:            time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
		ManagerLogin:        "manager1",
		WorkerLogin:         "worker1",
		CustomerLogin:       "customer1",
		PercentOfComplition: float64(1) / 3,
		PriceUnfinished:     5000,
		PriceTotal:          10000,
		Status:              2,
		Tasks:               mockTasks,
	}
	assert.Equal(t, expectedOrder, gotOrder)
}

func TestCreateOrder(t *testing.T) {
	orderRepoMock := new(orderRepoMock)
	taskRepoMock := new(taskRepoMock)
	testService := service.NewCustomerService(orderRepoMock, taskRepoMock, nil, nil)

	mockTasks := []model.TaskCreationDTO{
		{
			ItemID:    1,
			Name:      "item1",
			ItemPrice: 10,
			Amount:    20,
		},
		{
			ItemID:    2,
			Name:      "item2",
			ItemPrice: 30,
			Amount:    10,
		},
	}

	mockOrder := model.OrderCreationDTO{
		Deadline: time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
		Tasks:    mockTasks,
	}

	tasksToSave := []model.Task{
		{
			Name:     "Задача: изготовление item1 до 01.04.2026",
			OrderID:  1,
			ItemID:   1,
			Amount:   20,
			Finished: false,
			Price:    200,
		},
		{
			Name:     "Задача: изготовление item2 до 01.04.2026",
			OrderID:  1,
			ItemID:   2,
			Amount:   10,
			Finished: false,
			Price:    300,
		},
	}
	orderToSave := model.Order{
		Name:          "Заказ для пользователя customer1 до 01.04.2026 на сумму 5.0",
		Deadline:      time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
		CustomerLogin: "customer1",
		PriceTotal:    500,
		Status:        0,
	}

	orderRepoMock.On("SaveOrder", orderToSave).Return(1, nil)

	taskRepoMock.On("SaveTask", tasksToSave[0]).Return(1, nil)
	taskRepoMock.On("SaveTask", tasksToSave[1]).Return(2, nil)

	id, saveErr := testService.CreateOrder("customer1", 1, mockOrder)
	assert.NoError(t, saveErr)
	assert.Equal(t, 1, id)
	assert.True(t, taskRepoMock.tx.CommitCalled)
}

func TestDeleteOrder(t *testing.T) {
	orderRepoMock := new(orderRepoMock)
	taskRepoMock := new(taskRepoMock)
	testService := service.NewCustomerService(orderRepoMock, taskRepoMock, nil, nil)

	mockOrder := model.Order{
		ID:            1,
		Name:          "order1",
		Deadline:      time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC),
		ManagerLogin:  "manager1",
		WorkerLogin:   "worker1",
		CustomerLogin: "customer1",
		PriceTotal:    5000,
		Status:        1,
	}

	mockTasks := []model.Task{
		{
			ID:       1,
			Name:     "task1",
			OrderID:  1,
			ItemID:   1,
			Amount:   1,
			Finished: true,
			Price:    3000,
		},
		{
			ID:       2,
			Name:     "task2",
			OrderID:  1,
			ItemID:   2,
			Amount:   1,
			Finished: false,
			Price:    2000,
		},
	}

	orderRepoMock.On("GetOrderById", 1).Return(mockOrder, nil)
	orderRepoMock.On("DeleteOrder", 1).Return(nil)

	taskRepoMock.On("GetTasksByContract", 1).Return(mockTasks, nil)
	taskRepoMock.On("DeleteTask", 1).Return(nil)
	taskRepoMock.On("DeleteTask", 2).Return(nil)

	deletedId, deleteErr := testService.DeleteOrder("customer1", 1, 1)
	assert.NoError(t, deleteErr)
	assert.Equal(t, 2000, deletedId)
	assert.True(t, taskRepoMock.tx.CommitCalled)
}

func TestGetItems(t *testing.T) {
	itemRepoMock := new(itemRepoMock)
	testService := service.NewCustomerService(nil, nil, nil, itemRepoMock)

	mockItems := []model.Item{
		{ID: 1, Name: "item1", Price: 10},
		{ID: 2, Name: "item2", Price: 20},
	}

	itemRepoMock.On("GetAllItems").Return(mockItems, nil)

	gotItems, itemsErr := testService.GetItems("customer1", 1)
	assert.NoError(t, itemsErr)
	assert.Equal(t, mockItems, gotItems)
}

func TestGetBasket(t *testing.T) {
	basketRepoMock := new(basketRepoMock)
	testService := service.NewCustomerService(nil, nil, basketRepoMock, nil)

	mockItems := []model.TaskCreationDTO{
		{ItemID: 1, Name: "item1", ItemPrice: 10, Amount: 2},
		{ItemID: 2, Name: "item2", ItemPrice: 20, Amount: 1},
	}

	basketRepoMock.On("GetBasket", "customer1").Return(mockItems, nil)

	gotBasket, basketErr := testService.GetBasket("customer1", 1)
	assert.NoError(t, basketErr)
	assert.Equal(t, mockItems, gotBasket)
}

func TestSaveToBasket(t *testing.T) {
	basketRepoMock := new(basketRepoMock)
	testService := service.NewCustomerService(nil, nil, basketRepoMock, nil)

	mockTask := model.TaskCreationDTO{ItemID: 1, Name: "item1", ItemPrice: 10, Amount: 2}

	basketRepoMock.On("SaveToBasket", "customer1", mockTask).Return(nil)
	saveErr := testService.SaveToBasket("customer1", 1, mockTask)
	assert.NoError(t, saveErr)
}

func TestDeleteFromBasket(t *testing.T) {
	basketRepoMock := new(basketRepoMock)
	testService := service.NewCustomerService(nil, nil, basketRepoMock, nil)

	basketRepoMock.On("DeleteFromBasket", "customer1", 1).Return(nil)
	deleteErr := testService.DeleteFromBasket("customer1", 1, 1)
	assert.NoError(t, deleteErr)
}

func TestClearBasket(t *testing.T) {
	basketRepoMock := new(basketRepoMock)
	testService := service.NewCustomerService(nil, nil, basketRepoMock, nil)
	basketRepoMock.On("ClearBasket", "customer1").Return(nil)

	testService.ClearBasket("customer1", 1)
}
