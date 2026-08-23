package service_test

import (
	"context"
	"crud-go/internal/model"
	"crud-go/internal/repository"
	"database/sql"
	"time"

	"github.com/stretchr/testify/mock"
)

// BasketRepoMock
type basketRepoMock struct {
	mock.Mock
	tx *MockTx
}

func (m *basketRepoMock) GetBasket(login string) ([]model.TaskCreationDTO, error) {
	args := m.Called(login)
	return args.Get(0).([]model.TaskCreationDTO), args.Error(1)
}

func (m *basketRepoMock) SaveToBasket(login string, item model.TaskCreationDTO) error {
	args := m.Called(login, item)
	return args.Error(0)
}

func (m *basketRepoMock) DeleteFromBasket(login string, itemID int) error {
	args := m.Called(login, itemID)
	return args.Error(0)
}

func (m *basketRepoMock) ClearBasket(login string) error {
	args := m.Called(login)
	return args.Error(0)
}

func (m *basketRepoMock) Begin() (repository.Tx, error) {
	m.tx = new(MockTx)
	return m.tx, nil
}

func (m *basketRepoMock) BeginTx(ctx context.Context, opts *sql.TxOptions) (repository.Tx, error) {
	m.tx = new(MockTx)
	return m.tx, nil
}

// ItemRepoMock
type itemRepoMock struct {
	mock.Mock
}

func (m *itemRepoMock) GetAllItems() ([]model.Item, error) {
	args := m.Called()
	return args.Get(0).([]model.Item), args.Error(1)
}

func (m *itemRepoMock) GetItemById(id int) (model.Item, error) {
	args := m.Called(id)
	return args.Get(0).(model.Item), args.Error(1)
}

// OrderRepoMock
type orderRepoMock struct {
	mock.Mock
}

func (m *orderRepoMock) GetAllOrders() ([]model.Order, error) {
	args := m.Called()
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *orderRepoMock) GetOrdersByManager(login string) ([]model.Order, error) {
	args := m.Called(login)
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *orderRepoMock) GetOrdersByWorker(login string) ([]model.Order, error) {
	args := m.Called(login)
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *orderRepoMock) GetOrdersByCustomer(login string) ([]model.Order, error) {
	args := m.Called(login)
	return args.Get(0).([]model.Order), args.Error(1)
}

func (m *orderRepoMock) GetOrderById(id int) (model.Order, error) {
	args := m.Called(id)
	return args.Get(0).(model.Order), args.Error(1)
}

func (m *orderRepoMock) SaveOrder(order model.Order) (int, error) {
	args := m.Called(order)
	return args.Int(0), args.Error(1)
}

func (m *orderRepoMock) DeleteOrder(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

// TaskRepoMock
type taskRepoMock struct {
	mock.Mock
	tx *MockTx
}

func (m *taskRepoMock) GetAllTasks() ([]model.Task, error) {
	args := m.Called()
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *taskRepoMock) GetTasksByContracts(contractID []int) ([]model.Task, error) {
	args := m.Called(contractID)
	return args.Get(0).([]model.Task), args.Error(1)
}

func (m *taskRepoMock) GetTaskById(id int) (model.Task, error) {
	args := m.Called(id)
	return args.Get(0).(model.Task), args.Error(1)
}

func (m *taskRepoMock) SaveTask(task model.Task) (int, error) {
	args := m.Called(task)
	return args.Int(0), args.Error(1)
}

func (m *taskRepoMock) DeleteTask(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *taskRepoMock) Begin() (repository.Tx, error) {
	m.tx = new(MockTx)
	return m.tx, nil
}

func (m *taskRepoMock) BeginTx(ctx context.Context, opts *sql.TxOptions) (repository.Tx, error) {
	m.tx = new(MockTx)
	return m.tx, nil
}

// UserRepoMock
type userRepoMock struct {
	mock.Mock
}

func (m *userRepoMock) GetWorkersByManager(managerLogin string) ([]model.Worker, error) {
	args := m.Called(managerLogin)
	return args.Get(0).([]model.Worker), args.Error(1)
}

func (m *userRepoMock) GetRoleByLogin(login string) (int, string, error) {
	args := m.Called(login)
	return args.Int(0), args.String(1), args.Error(2)
}

func (m *userRepoMock) GetWorker(workerLogin string) (model.Worker, error) {
	args := m.Called(workerLogin)
	return args.Get(0).(model.Worker), args.Error(1)
}

func (m *userRepoMock) SaveCustomer(customer model.Customer) error {
	args := m.Called(customer)
	return args.Error(0)
}

// TokenRepoMock
type tokenRepoMock struct {
	mock.Mock
}

func (m *tokenRepoMock) Save(tokenID, login, tokenHash string, expiresAt time.Time) error {
	args := m.Called(tokenID, login, tokenHash, expiresAt)
	return args.Error(0)
}

func (m *tokenRepoMock) GetTokenByLogin(login string) (model.RefreshToken, error) {
	args := m.Called(login)
	return args.Get(0).(model.RefreshToken), args.Error(1)
}

func (m *tokenRepoMock) Revoke(login string) error {
	args := m.Called(login)
	return args.Error(0)
}

// Transaction Mock
type MockTx struct {
	CommitCalled bool
}

func (m *MockTx) Commit() error {
	m.CommitCalled = true
	return nil
}

func (m *MockTx) Rollback() error {
	return nil
}
