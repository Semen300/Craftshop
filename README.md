# Craftshop

`Craftshop` — это REST API на Go для магазинов, реализующих собственную продукцию с ролью клиента, рабочего и менеджера. Проект использует Gin, PostgreSQL и JWT-авторизацию.

## Основные особенности

- Приложение построено по монолитной архитектуре
- Возможность регистрации и аутентификации пользователей
- JWT access / refresh токены
- Отдельные роли: customer, worker, manager
- Управление заказами, корзиной, задачами и товарами
- Настраиваемая логика работы с заказами и назначением исполнителя
- Тестирование с использованием `testify` и `testcontainers`

## Технологии

- Go 1.26
- Gin Web Framework
- PostgreSQL (pgx)
- JWT
- dotenv

## Архитектура

Приложение построено на трёхуровневой "handler-service-repository" архитектуре. Список основных директорий проекта:

- `cmd/app/main.go` - точка входа приложения.
- `internal/handler` - HTTP-обработчики.
- `internal/service` - бизнес-логика.
- `internal/repository` - подключение к БД и CRUD операции.
- `internal/middleware` - промежуточная авторизация.
- `internal/model` - модели данных.

В каждой директории расположен одноимённый пакет и, где это необходимо, пакет с тестами.

## Требования

- Go 1.26+
- PostgreSQL
- Переменные окружения, описанные ниже

## Переменные окружения

Пример `.env`:

```env
APP_PORT=8080
DATABASE_USERNAME=postgres
DATABASE_PASSWORD=password
DATABASE_PORT=5432
DATABASE_SCHEMA=crud_db
AUTH_ACCESS_LIFETIME=300 //Время жизни access-токена в секундах
AUTH_REFRESH_LIFETIME=86400 //Время жизни refresh-токена в секундах
AUTH_ACCESS_KEY=your_access_secret
AUTH_REFRESH_KEY=your_refresh_secret
```

## Запуск

1. Установите PostgreSQL.
2. Создайте базу данных, указанную в `DATABASE_SCEMA`.
3. Создайте `.env` с нужными переменными.
4. Запустите приложение:

```bash
go run ./cmd/app
```

## Тесты

Тесты запускаются при помощи команды

```bash
go test ./...
```

Реализовано покрытие 75% критического функционала Unit- и интеграционными тестами.

## API


### Версионирование

Актуальная версия API - `v1`, все запросы должны иметь префикс `/api/v1` к базовому URL.


### Авторизация

#### **Вход в систему:**

Данный эндпоинт используется для входа пользователя в систему и получения им токенов доступа. Пользователь должен передать свой логин и пароль в теле запроса.

*Метод:* `POST`  
*URL:* `/auth/login`  
*Пример тела запроса:*  
```json
{
    "login":"login",
    "password":"1234"
}
```
*Пример тела ответа:*
```json
{
    {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6InNpbW9uIiwicm9sZSI6MSwiZXhwIjoxNzkzMjc0MjcxfQ.zy_-SzkG6NJfLvHdw70x6gA4vY8ye-T7DCm67Ss3WgQ",
    "claims": {
        "login": "simon",
        "role": 1
    },
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6InNpbW9uIiwicm9sZSI6MSwiZXhwIjoxNzkzMzU3NjcxLCJqdGkiOiIwZDk5Y2YzOS05NmFiLTRkZDYtOTk2OC02YmU4MDZiNDA1MmMifQ.fypSMo3u4cgB_1SarD6qB6BOkB6hCKtmLw67ZZYooC8"
}
}
```

#### **Обновление access token**

Эндпоинт предназначен для обновления access токена по действующему refresh токену. Необходимо передать refresh токен в теле запроса.

*Метод:* `POST`  
*URL:* `/auth/refresh`  
*Пример тела запроса:*
```json
{
    "refreshToken":"sdetgdwr.strrtggwr.s32tgso5"
}
```
*Пример тела ответа:*
```json
{
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJsb2dpbiI6InNpbW9uIiwicm9sZSI6MSwiZXhwIjoxNzkzMjc0OTI4LCJqdGkiOiIwZDk5Y2YzOS05NmFiLTRkZDYtOTk2OC02YmU4MDZiNDA1MmMifQ.4BYI_ZkGxkK_CBI3_7DDtZbln76hi_rf_uXloGJ67C0"
}
```

#### **Выход из системы**

Эндпоинт предназначен для выхода пользователя из системы и отзыва его refresh токена. Пользователь должен передать свой логин в теле запроса.

*Метод:* `POST`  
*URL:* `/auth/logout`  
*Пример тела запроса:*
```json
{
    "login":"login"
}
```
*Пример тела ответа:*
```json
{
    "message": "Successfully logged out"
}
```


### Функционал клиента

Все представленные ниже эндпоинты, кроме эндпоинта для регистрации нового клиента, доступны для пользователя после его входа в систему в качестве клиента.

#### **Регистриция нового клиента**

Эндпоинт предназначен для регистрации нового покупателя. Пользователь должен передать информацию для регистрации (логин, пароль, ФИО, номер телефона и почту).

*Метод:* `POST`  
*URL:* `/customer`  
*Пример тела запроса:* 
```json
{
    "login":"user",
    "password":"1234",
    "fio":"Smirnov S.V.",
    "number":"+78005553535",
    "email":"test_mail@gmail.com"
}
``` 
*Пример тела ответа:*
```json
{
    "message": "Registration successful"
}
```

#### **Получение списка заказов**

Эндпоинт предназначен для получения списка заказов пользователя.

*Метод:* `GET`  
*URL:* `/customer/orders`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
[
    [
    {
        "ID": 1,
        "Name": "Заказ для пользователя simon до 01.04.2030 на сумму 5.0",
        "Deadline": "2030-04-01T00:00:00Z",
        "ManagerLogin": "",
        "WorkerLogin": "",
        "CustomerLogin": "simon",
        "PercentOfComplition": 0,
        "PriceTotal": 500,
        "PriceUnfinished": 500,
        "Status": 0,
        "Tasks": [
            {
                "ID": 1,
                "Name": "Задача: изготовление Item1 до 01.04.2030",
                "OrderID": 1,
                "ItemID": 1,
                "Amount": 1,
                "Finished": false,
                "Price": 100
            },
            {
                "ID": 2,
                "Name": "Задача: изготовление Item2 до 01.04.2030",
                "OrderID": 1,
                "ItemID": 2,
                "Amount": 2,
                "Finished": false,
                "Price": 400
            }
        ]
    }
]
```


#### **Получение информации о заказе**

Эндпоинт предназначен для получения информации о заказе по его id. Пользователь должен пердварительно выполнить вход в систему.

*Метод:* `GET`  
*URL:* `/customer/orders/{id}`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
{
    "ID": 1,
    "Name": "Заказ для пользователя simon до 01.04.2030 на сумму 5.0",
    "Deadline": "2030-04-01T00:00:00Z",
    "ManagerLogin": "",
    "WorkerLogin": "",
    "CustomerLogin": "simon",
    "PercentOfComplition": 0,
    "PriceTotal": 500,
    "PriceUnfinished": 500,
    "Status": 0,
    "Tasks": [
        {
            "ID": 1,
            "Name": "Задача: изготовление Item1 до 01.04.2030",
            "OrderID": 1,
            "ItemID": 1,
            "Amount": 1,
            "Finished": false,
            "Price": 100
        },
        {
            "ID": 2,
            "Name": "Задача: изготовление Item2 до 01.04.2030",
            "OrderID": 1,
            "ItemID": 2,
            "Amount": 2,
            "Finished": false,
            "Price": 400
        }
    ]
}
```

#### **Создание нового заказа**

Эндпоинт предназначен для создания нового заказа на основе данных, переданных в теле запроса.

*Метод:* `POST`  
*URL:* `/customer/orders`  
*Пример тела запроса:* 
```json
{
    "deadline": "2030-04-01T00:00:00Z",
    "tasks": [
        {
            "itemId": 1,
            "name": "Item1",
            "itemPrice": 100,
            "amount": 1
        },
        {
            "itemId": 2,
            "name": "Item2",
            "itemPrice": 200,
            "amount": 2
        }
    ]
}
``` 
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Запрос не возвращает тело ответа*

#### **Удаление заказа**

Эндпоинт пердназначен для удаления заказа по его id.

*Метод:* `DELETE`  
*URL:* `/customer/orders/{id}`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Запрос не возвращает тело ответа*

#### **Получение списка товаров**

Эндпоинт предназначен для получения списка (ассортимента, каталога) всех товаров, которые реализает предприятие.

*Метод:* `GET`  
*URL:* `/customer/items`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
[
    {
        "ID": 1,
        "Name": "Item1",
        "Price": 100
    },
    {
        "ID": 2,
        "Name": "Item2",
        "Price": 200
    }
]
```

#### **Получение корзины**

Эндпоинт предназначен для получения пользователем своей корзины.

*Метод:* `GET`  
*URL:* `/customer/basket`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
[
    {
        "itemId": 1,
        "name": "Item1",
        "itemPrice": 1000,
        "amount": 20
    }
]
```

#### **Добавление товара в корзину**

Эндпоинт предназначен для добавления товара в корзину или изменения уже существующего товара. Если в корзине пользователя нет товара с данным ID, то происходит добавление нового, иначе - изменение товара в корзине в соответствии с данными, переданными в теле запроса.

*Метод:* `POST`  
*URL:* `/customer/basket`  
*Пример тела запроса:* 
```json
{
    "itemId": 1,
    "name": "Item1",
    "itemPrice": 100,
    "amount": 1
}
```
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Запрос не возвращает тело ответа*

#### **Удаление товара из корзины**

Эндпоинт предназначен для удаления товара из корзины.

*Метод:* `DELETE`  
*URL:* `/customer/basket/{id}`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Запрос не возвращает тело ответа*


### Функционал работника

Все представленные ниже эндпоинты доступны для пользователя после его входа в систему в качестве работника.

#### **Просмотр списка заказов**

Эндпоинт предназначен для получения работником списка назначенных ему заказов.

*Метод:* `GET`  
*URL:* `/worker/orders`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
[
    {
        "ID": 4,
        "Name": "Заказ для пользователя simon до 01.04.2030 на сумму 5.0",
        "Deadline": "2030-04-01T00:00:00Z",
        "ManagerLogin": "manager1",
        "WorkerLogin": "worker1",
        "CustomerLogin": "simon",
        "PercentOfComplition": 0,
        "PriceTotal": 500,
        "PriceUnfinished": 500,
        "Status": 1,
        "Tasks": [
            {
                "ID": 7,
                "Name": "Задача: изготовление Item1 до 01.04.2030",
                "OrderID": 4,
                "ItemID": 1,
                "Amount": 1,
                "Finished": false,
                "Price": 100
            },
            {
                "ID": 8,
                "Name": "Задача: изготовление Item2 до 01.04.2030",
                "OrderID": 4,
                "ItemID": 2,
                "Amount": 2,
                "Finished": false,
                "Price": 400
            }
        ]
    }
]
```

#### **Просмотр заказа рабочего по его ID**

Эндпоинт предназначен для получения работником заказа по ID заказа.

*Метод:* `GET`  
*URL:* `/worker/orders/{id}`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
{
    "ID": 4,
    "Name": "Заказ для пользователя simon до 01.04.2030 на сумму 5.0",
    "Deadline": "2030-04-01T00:00:00Z",
    "ManagerLogin": "manager1",
    "WorkerLogin": "worker1",
    "CustomerLogin": "simon",
    "PercentOfComplition": 0,
    "PriceTotal": 500,
    "PriceUnfinished": 500,
    "Status": 1,
    "Tasks": [
        {
            "ID": 7,
            "Name": "Задача: изготовление Item1 до 01.04.2030",
            "OrderID": 4,
            "ItemID": 1,
            "Amount": 1,
            "Finished": false,
            "Price": 100
        },
        {
            "ID": 8,
            "Name": "Задача: изготовление Item2 до 01.04.2030",
            "OrderID": 4,
            "ItemID": 2,
            "Amount": 2,
            "Finished": false,
            "Price": 400
        }
    ]
}
```

#### **Отметка о выполнении задачи**

Эндпоинт предназначен для отметки работником задачи как выполненной.

*Метод:* `POST`  
*URL:* `/worker/tasks/{id}`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Запрос не возвращает тело ответа*


### Функционал менеджера

Все представленные ниже эндпоинты доступны для пользователя после его входа в систему в качестве менеджера.

#### **Просмотр списка заказов менеджера**

Эндпоинт предназначен для получения менеджером списка заказов, на которые ещё не назначены работники или назначены работники, подчинённые данному менеджеру.

*Метод:* `GET`  
*URL:* `/manager/orders`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
[
    {
        "ID": 2,
        "Name": "Заказ для пользователя simon до 01.04.2030 на сумму 5.0",
        "Deadline": "2030-04-01T00:00:00Z",
        "ManagerLogin": "",
        "WorkerLogin": "",
        "CustomerLogin": "simon",
        "PercentOfComplition": 0,
        "PriceTotal": 500,
        "PriceUnfinished": 500,
        "Status": 0,
        "Tasks": [
            {
                "ID": 3,
                "Name": "Задача: изготовление Item1 до 01.04.2030",
                "OrderID": 2,
                "ItemID": 1,
                "Amount": 1,
                "Finished": false,
                "Price": 100
            },
            {
                "ID": 4,
                "Name": "Задача: изготовление Item2 до 01.04.2030",
                "OrderID": 2,
                "ItemID": 2,
                "Amount": 2,
                "Finished": false,
                "Price": 400
            }
        ]
    },
    {
        "ID": 3,
        "Name": "Заказ для пользователя simon до 01.04.2030 на сумму 5.0",
        "Deadline": "2030-04-01T00:00:00Z",
        "ManagerLogin": "",
        "WorkerLogin": "",
        "CustomerLogin": "simon",
        "PercentOfComplition": 0,
        "PriceTotal": 500,
        "PriceUnfinished": 500,
        "Status": 0,
        "Tasks": [
            {
                "ID": 5,
                "Name": "Задача: изготовление Item1 до 01.04.2030",
                "OrderID": 3,
                "ItemID": 1,
                "Amount": 1,
                "Finished": false,
                "Price": 100
            },
            {
                "ID": 6,
                "Name": "Задача: изготовление Item2 до 01.04.2030",
                "OrderID": 3,
                "ItemID": 2,
                "Amount": 2,
                "Finished": false,
                "Price": 400
            }
        ]
    }
]
```

#### **Получение заказа по его ID**

Эндпоинт предназначен для получения менеджером заказа по ID заказа.

*Метод:* `GET`  
*URL:* `/manager/orders/{id}`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
{
    "ID": 3,
    "Name": "Заказ для пользователя simon до 01.04.2030 на сумму 5.0",
    "Deadline": "2030-04-01T00:00:00Z",
    "ManagerLogin": "",
    "WorkerLogin": "",
    "CustomerLogin": "simon",
    "PercentOfComplition": 0,
    "PriceTotal": 500,
    "PriceUnfinished": 500,
    "Status": 0,
    "Tasks": [
        {
            "ID": 5,
            "Name": "Задача: изготовление Item1 до 01.04.2030",
            "OrderID": 3,
            "ItemID": 1,
            "Amount": 1,
            "Finished": false,
            "Price": 100
        },
        {
            "ID": 6,
            "Name": "Задача: изготовление Item2 до 01.04.2030",
            "OrderID": 3,
            "ItemID": 2,
            "Amount": 2,
            "Finished": false,
            "Price": 400
        }
    ]
}
```

#### **Назначение работника на заказ**

Эндпоинт предназначен для назначения менеджером работника на заказ.

*Метод:* `POST`  
*URL:* `/manager/orders/{id}`  
*Пример тела запроса:* 
```json
{
    "workerLogin":"worker1"
}
```
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Запрос не возвращает тело ответа*

#### **Получение списка рабочих**

Эндпоинт предназначен для получения менеджером списка подчинённых ему рабочих.

*Метод:* `GET`  
*URL:* `/manager/workers`  
*Запрос выполняется с пустым телом*  
*Обязательные заголовки запроса:* `Autorization: Bearer <access token>`
*Пример тела ответа:*
```json
[
    {
        "Login": "worker1",
        "Password": "sduiwk21fs",
        "Fio": "Worker",
        "SuperiorLogin": "manager1"
    }
]
```