package service

import (
	"crud-go/internal/model"
	"crud-go/internal/repository"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type IAuthService interface {
	RegisterNewCustomer(string, string, string, string, string) error
	Login(string, string) (model.Claims, string, string, error)
	Refresh(string) (string, error)
	Logout(string) error
	GenerateAccessToken(model.Claims) (string, error)
	GenerateRefreshToken(model.Claims) (string, string, time.Time, error)
	ParseAccessToken(string) (model.Claims, error)
	ParseRefreshToken(string) (model.Claims, error)
}

type AuthService struct {
	UserRepo        *repository.IUserRepository
	TokenRepo       *repository.ITokenRepository
	accessLifetime  int
	refreshLifetime int
	accessKey       string
	refreshKey      string
}

// NewAuthService создаёт новый сервис для выполнения операций аутентификации.
//
// Принимает интерфейсы репозиториев пользователей и токенов, время жизни токенов
// и ключи для их подписи,
// возвращает новый экземпляр сервиса.
func NewAuthService(ur repository.IUserRepository, tr repository.ITokenRepository, accessLifetime, refreshLifetime int, accessKey, refreshKey string) AuthService {
	return AuthService{
		UserRepo:        &ur,
		TokenRepo:       &tr,
		accessLifetime:  accessLifetime,
		refreshLifetime: refreshLifetime,
		accessKey:       accessKey,
		refreshKey:      refreshKey,
	}
}

// RegisterNewCustomer служит для регистрации нового покупателя.
//
// Принимает логин, пароль, ФИО, номер телефона и адрес электронной почты покупателя,
// возвращает возможную ошибку.
func (as AuthService) RegisterNewCustomer(login, password, fio, number, email string) error {
	ur := *as.UserRepo
	hashedPass := CreateHash(password)
	customerToSave := model.NewCustomer(login, hashedPass, fio, number, email)
	saveErr := ur.SaveCustomer(customerToSave)
	if saveErr != nil {
		return fmt.Errorf("Error creating new customer: \n%w", saveErr)
	}
	return nil
}

// Login служит для аутентификации пользователя и создания токенов доступа.
//
// Принимает логин и пароль пользователя,
// возвращает данные пользователя, токен доступа, токен обновления и возможную ошибку.
func (as AuthService) Login(login, password string) (model.Claims, string, string, error) {
	ur := *as.UserRepo
	tr := *as.TokenRepo
	role, savedPassword, err := ur.GetRoleByLogin(login)
	if err != nil {
		return model.Claims{}, "", "", err
	}
	hashedPass := CreateHash(password)
	if hashedPass != savedPassword {
		return model.Claims{}, "", "", fmt.Errorf("Incorrect login or password")
	}

	accessToken, atErr := as.GenerateAccessToken(
		model.Claims{
			Login: login,
			Role:  role,
		},
	)
	if atErr != nil {
		return model.Claims{}, "", "", atErr
	}

	rTokenId, refreshToken, expirationTime, rtErr := as.GenerateRefreshToken(
		model.Claims{
			Login: login,
			Role:  role,
		},
	)
	if rtErr != nil {
		return model.Claims{}, "", "", rtErr
	}

	revokeErr := tr.Revoke(login)
	if revokeErr != nil {
		return model.Claims{}, "", "", revokeErr
	}

	hashedToken := CreateHash(refreshToken)

	saveErr := tr.Save(rTokenId, login, hashedToken, expirationTime)
	if saveErr != nil {
		return model.Claims{}, "", "", saveErr
	}

	return model.Claims{
			Login: login,
			Role:  role,
		},
		accessToken,
		refreshToken,
		nil
}

// Refresh служит для обновления токена доступа с помощью токена обновления.
//
// Принимает строковое представление токена обновления,
// возвращает новый токен доступа и возможную ошибку.
func (as AuthService) Refresh(refreshTokenStr string) (string, error) {
	tr := *as.TokenRepo
	claims, parseErr := as.ParseRefreshToken(refreshTokenStr)
	if parseErr != nil {
		return "", parseErr
	}
	refreshToken, getErr := tr.GetTokenByLogin(claims.Login)
	if getErr != nil {
		return "", getErr
	}

	if refreshToken.Revoked || refreshToken.ExpiresAt.Before(time.Now()) || CreateHash(refreshTokenStr) != refreshToken.TokenHash {
		return "", fmt.Errorf("Error: refresh token invalid\n")
	}

	return as.GenerateAccessToken(claims)
}

// Logout служит для завершения сессии пользователя.
//
// Принимает логин пользователя,
// возвращает возможную ошибку.
func (as AuthService) Logout(login string) error {
	tr := *as.TokenRepo
	return tr.Revoke(login)
}

// GenerateAccessToken служит для создания токена доступа.
//
// Принимает данные пользователя для включения в токен,
// возвращает строковое представление токена доступа и возможную ошибку.
func (as AuthService) GenerateAccessToken(claims model.Claims) (string, error) {
	expirationTime := time.Now().Add(time.Duration(as.accessLifetime) * time.Second)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(as.accessKey))
}

// GenerateRefreshToken служит для создания токена обновления.
//
// Принимает данные пользователя для включения в токен,
// возвращает ID токена, его строковое представление, время истечения действия и возможную ошибку.
func (as AuthService) GenerateRefreshToken(claims model.Claims) (string, string, time.Time, error) {
	expirationTime := time.Now().Add(time.Duration(as.refreshLifetime) * time.Second)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)

	tokenID := uuid.NewString()
	claims.ID = tokenID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, tokenErr := token.SignedString([]byte(as.refreshKey))
	return tokenID, tokenStr, expirationTime, tokenErr
}

// ParseAccessToken служит для проверки и разбора токена доступа.
//
// Принимает строковое представление токена доступа,
// возвращает данные пользователя из токена и возможную ошибку.
func (as AuthService) ParseAccessToken(tokenStr string) (model.Claims, error) {
	token, parseErr := jwt.ParseWithClaims(
		tokenStr,
		&model.Claims{},
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(as.accessKey), nil
		},
	)
	if parseErr != nil {
		return model.Claims{}, parseErr
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok {
		return model.Claims{}, jwt.ErrTokenInvalidClaims
	}

	if !token.Valid {
		return model.Claims{}, jwt.ErrTokenMalformed
	}

	return *claims, nil
}

// ParseRefreshToken служит для проверки и разбора токена обновления.
//
// Принимает строковое представление токена обновления,
// возвращает данные пользователя из токена и возможную ошибку.
func (as AuthService) ParseRefreshToken(tokenStr string) (model.Claims, error) {
	token, parseErr := jwt.ParseWithClaims(
		tokenStr,
		&model.Claims{},
		func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(as.refreshKey), nil
		})

	if parseErr != nil {
		return model.Claims{}, parseErr
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok {
		return model.Claims{}, jwt.ErrTokenInvalidClaims
	}

	if !token.Valid {
		return model.Claims{}, jwt.ErrTokenMalformed
	}

	return *claims, nil
}
