package handler

import (
	"net/http"
	"strconv"

	"crud-go/internal/model/errors"
	"crud-go/internal/service"

	"github.com/gin-gonic/gin"
)

type WorkerHandler struct {
	workerService service.IWorkerService
}

func NewWorkerHandler(workerService service.IWorkerService) WorkerHandler {
	return WorkerHandler{workerService: workerService}
}

// GetOrdersByWorker предназначена для получения всех заказов, назначенных рабочему.
//
// Принимает контекст запроса и сервис для работы с рабочими,
// возвращает список всех заказов, связанных с рабочим, который сделал запрос.
func (wh *WorkerHandler) GetOrdersByWorker(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	orders, serviceErr := wh.workerService.GetAllOrders(login, role)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving orders: " + serviceErr.Error()})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// GetOrder предназначена для получения информации о конкретном заказе, назначенном рабочему.
//
// Принимает контекст запроса и сервис для работы с рабочими,
// возвращает информацию о заказе с указанным ID.
func (wh WorkerHandler) GetOrder(ctx *gin.Context) {
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

	order, serviceErr := wh.workerService.GetOrderById(login, role, orderID)
	if serviceErr != nil {
		switch serviceErr.(type) {
		case errors.AccessRejectedError:
			{
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": serviceErr.Error()})
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving order: " + serviceErr.Error()})
		return
	}

	ctx.JSON(http.StatusOK, order)
}

// SetTaskCompleted предназначена для установки статуса задачи как выполненной.
//
// Принимает контекст запроса и сервис для работы с рабочими,
// устанавливает статус задачи как выполненной по ID задачи, указанному в URL, для рабочего, который сделал запрос.
// В случае успеха возвращает статус 204 No Content и URL задачи в заголовке Location, иначе - сообщение об ошибке.
func (wh WorkerHandler) SetTaskCompleted(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	taskID, paramErr := strconv.Atoi(ctx.Param("id"))
	if paramErr != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	serviceErr := wh.workerService.SetTaskCompleted(login, role, taskID)
	if serviceErr != nil {
		switch serviceErr.(type) {
		case errors.AccessRejectedError:
			{
				ctx.AbortWithStatusJSON(http.StatusForbidden, serviceErr.Error())
				return
			}
		default:
			{
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error setting task as completed: " + serviceErr.Error()})
				return
			}
		}
	}

	ctx.Header("Location", "/tasks/"+strconv.Itoa(taskID))
	ctx.Status(http.StatusNoContent)
}
