package handler

import (
	"crud-go/internal/model"
	"crud-go/internal/model/errors"
	"crud-go/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CustomerHandler предназначен для обработки HTTP-запросов, связанных с заказами и товарами клиентов.
type CustomerHandler struct {
	CustomerService service.ICustomerService
}

// NewCustomerHandler создаёт новый обработчик для работы с заказами и товарами клиентов.
//
// Принимает сервис для работы с клиентами,
// возвращает новый экземпляр обработчика.
func NewCustomerHandler(customerService service.ICustomerService) CustomerHandler {
	return CustomerHandler{
		CustomerService: customerService,
	}
}

// getUserInfo служит для извлечения логина и роли пользователя из контекста запроса.
//
// Принимает контекст запроса, возвращает логин и роль пользователя.
// Если логин или роль отсутствуют, или если возникает ошибка при их извлечении, функция отправляет соответствующий HTTP-ответ и возвращает пустые значения.
func getUserInfo(ctx *gin.Context) (string, int) {

	login, loginExists := ctx.Params.Get("login")
	role, roleExists := ctx.Params.Get("role")

	if !loginExists || !roleExists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return "", 0
	}

	roleInt, _ := strconv.Atoi(role)
	return login, roleInt
}

// GetAllOrders предназначена для получения всех заказов клиента.
//
// Принимает контекст запроса и сервис для работы с клиентами,
// возвращает список всех заказов, связанных с клиентом, который сделал запрос.
func (ch CustomerHandler) GetAllOrders(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}

	orders, serviceErr := ch.CustomerService.GetOrdersByCustomer(login, role)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving orders: " + serviceErr.Error()})
		return
	}

	ctx.JSON(http.StatusOK, orders)
}

// GetCustomerOrder предназначена для получения конкретного заказа клиента по его ID.
//
// Принимает контекст запроса и сервис для работы с клиентами,
// возвращает заказ, связанный с клиентом, который сделал запрос, и имеющий указанный ID.
func (ch CustomerHandler) GetCustomerOrder(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}

	id, paramErr := strconv.Atoi(ctx.Param("id"))
	if paramErr != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	order, serviceErr := ch.CustomerService.GetOrderById(login, role, id)
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

// CreateOrder предназначена для создания нового заказа клиента.
//
// Принимает контекст запроса и сервис для работы с клиентами,
// возвращает статус 201 Created и URL нового заказа в заголовке Location.
func (ch CustomerHandler) CreateOrder(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}

	var orderDTO model.OrderCreationDTO
	if err := ctx.ShouldBindJSON(&orderDTO); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}
	createdOrderId, serviceErr := ch.CustomerService.CreateOrder(login, role, orderDTO)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error creating order: " + serviceErr.Error()})
		return
	}

	ctx.Header("Location", "/customer/orders/"+strconv.Itoa(createdOrderId))
	ctx.Status(http.StatusCreated)
	ctx.Writer.WriteHeaderNow()
}

// DeleteOrder предназначена для удаления заказа клиента по его ID.
//
// Принимает контекст запроса и сервис для работы с клиентами,
// удаляет заказ, связанный с клиентом, который сделал запрос, и имеющий указанный ID.
func (ch CustomerHandler) DeleteOrder(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}

	id, paramErr := strconv.Atoi(ctx.Param("id"))
	if paramErr != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	_, serviceErr := ch.CustomerService.DeleteOrder(login, role, id)
	if serviceErr != nil {
		switch serviceErr.(type) {
		case errors.AccessRejectedError:
			{
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": serviceErr.Error()})
				return
			}
		default:
			{
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error deleting order: " + serviceErr.Error()})
				return
			}
		}
	}
	ctx.Status(http.StatusNoContent)
	ctx.Writer.WriteHeaderNow()
}

// GetItems предназначена для получения списка всех товаров (каталога).
//
// Принимает контекст запроса и сервис для работы с клиентами,
// возвращает список всех товаров, доступных для заказа.
func (ch CustomerHandler) GetItems(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	items, serviceErr := ch.CustomerService.GetItems(login, role)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving items: " + serviceErr.Error()})
		return
	}

	ctx.JSON(http.StatusOK, items)
}

// GetBasket предназначена для получения текущей корзины клиента.
//
// Принимает контекст запроса и сервис для работы с клиентами,
// возвращает текущую корзину клиента, который сделал запрос.
func (ch CustomerHandler) GetBasket(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	basket, serviceErr := ch.CustomerService.GetBasket(login, role)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving basket: " + serviceErr.Error()})
		return
	}
	ctx.JSON(http.StatusOK, basket)
}

// SaveToBasket предназначена для сохранения товара в корзину клиента.
//
// Принимает контекст запроса и сервис для работы с клиентами,
// сохраняет товар в корзину клиента, который сделал запрос, используя DTO товара, переданный в теле запроса.
func (ch CustomerHandler) SaveToBasket(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	var item model.TaskCreationDTO
	if err := ctx.ShouldBindJSON(&item); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid item data: " + err.Error()})
		return
	}
	serviceErr := ch.CustomerService.SaveToBasket(login, role, item)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error saving item to basket: " + serviceErr.Error()})
		return
	}

	ctx.Status(http.StatusOK)
	ctx.Writer.WriteHeaderNow()
}

// DeleteFromBasket предназначена для удаления товара из корзины клиента.
//
// Принимает контекст запроса и сервис для работы с клиентами,
// удаляет товар из корзины клиента, который сделал запрос, используя ID товара, переданный в параметрах URL.
func (ch CustomerHandler) DeleteFromBasket(ctx *gin.Context) {
	login, role := getUserInfo(ctx)
	if login == "" || role == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Incorrect user claims"})
		return
	}
	id, paramErr := strconv.Atoi(ctx.Param("id"))
	if paramErr != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}
	serviceErr := ch.CustomerService.DeleteFromBasket(login, role, id)
	if serviceErr != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error deleting item from basket: " + serviceErr.Error()})
		return
	}

	ctx.Status(http.StatusOK)
	ctx.Writer.WriteHeaderNow()
}
