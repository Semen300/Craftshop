package handler

import (
	"crud-go/internal/model/errors"
	"crud-go/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ManagerHandler struct {
	managerService service.IManagerService
}

// NewManagerHandler создает обработчик для работы с запросами менеджера.
//
// Принимает сервис менеджера и возвращает новый обработчик.
func NewManagerHandler(managerService service.IManagerService) ManagerHandler {
	return ManagerHandler{managerService: managerService}
}

// GetAllWorkers предназначена для получения всех рабочих.
//
// Принимает контекст запроса и сервис для работы с менеджерами,
// возвращает список всех рабочих или сообщение об ошибке.
func (mh ManagerHandler) GetAllWorkers(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	workers, workerErr := mh.managerService.GetAllWorkers(login, role)
	if workerErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving workers: " + workerErr.Error()})
	}

	ctx.JSON(http.StatusOK, workers)
}

// GetOrdersByManager предназначена для получения всех заказов менеджера.
//
// Принимает контекст запроса и сервис для работы с менеджерами,
// возвращает список всех заказов, связанных с менеджером, который сделал запрос.
func (mh ManagerHandler) GetOrdersByManager(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	orders, serviceErr := mh.managerService.GetAllOrders(login, role)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving orders: " + serviceErr.Error()})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// GetOrderByID предназначена для получения информации о конкретном заказе.
//
// Принимает контекст запроса и сервис для работы с менеджерами,
// возвращает информацию о заказе с указанным ID или сообщение об ошибке.
func (mh ManagerHandler) GetOrderByID(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	orderID, paramErr := strconv.Atoi(ctx.Param("id"))
	if paramErr != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, serviceErr := mh.managerService.GetOrderById(login, role, orderID)
	if serviceErr != nil {
		switch serviceErr.(type) {
		case errors.AccessRejectedError:
			{
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": serviceErr.Error()})
				return
			}
		default:
			{
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving order: " + serviceErr.Error()})
				return
			}
		}
	}
	ctx.JSON(http.StatusOK, order)
}

// AssignWorkerToOrder предназначена для назначения рабочего на заказ.
//
// Принимает контекст запроса и сервис для работы с менеджерами,
// назначает рабочего на заказ по ID заказа и логину рабочего.
// В случае успеха возвращает статус 200 OK, иначе - сообщение об ошибке.
func (mh ManagerHandler) AssignWorkerToOrder(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	orderID, paramErr := strconv.Atoi(ctx.Param("id"))
	if paramErr != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	var requestBody struct {
		WorkerLogin string `json:"workerLogin"`
	}
	if bindErr := ctx.ShouldBindJSON(&requestBody); bindErr != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + bindErr.Error()})
		return
	}
	serviceErr := mh.managerService.AssignWorkerToOrder(login, role, orderID, requestBody.WorkerLogin)
	if serviceErr != nil {
		switch serviceErr.(type) {
		case errors.AccessRejectedError:
			{
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": serviceErr.Error()})
				return
			}
		default:
			{
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error setting worker login: " + serviceErr.Error()})
				return
			}
		}
	}
	ctx.Status(http.StatusOK)
	ctx.Writer.WriteHeaderNow()
}
