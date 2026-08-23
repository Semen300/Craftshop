package repository

import (
	"crud-go/internal/model"
	"database/sql"
	"time"
)

type ITokenRepository interface {
	Save(string, string, string, time.Time) error
	GetTokenByLogin(string) (model.RefreshToken, error)
	Revoke(string) error
}

type TokenRepository struct {
	Conn *sql.DB
}

// NewTokenRepository создаёт новый репозиторий для доступа к функционалу токенов.
// Также проводит инициализацию таблицы "tokens".
//
// Принимает указатель на подключение к базе данных,
// возвращает новый экземпляр репозитория и возможную ошибку.
func NewTokenRepository(db *sql.DB) (TokenRepository, error) {
	query := `CREATE TABLE IF NOT EXISTS tokens (
	token_id UUID PRIMARY KEY,
    login TEXT,
    token_hash TEXT,
    expires_at TIMESTAMP,
    revoked BOOLEAN DEFAULT false
	)`

	_, createErr := db.Exec(query)
	if createErr != nil {
		return TokenRepository{}, createErr
	}

	return TokenRepository{Conn: db}, nil
}

// Save служит для сохранения токена в базе данных.
//
// Принимает ID токена, логин пользователя, хеш токена и время истечения его действия,
// возвращает возможную ошибку.
func (tr TokenRepository) Save(tokenID, login, tokenHash string, expiresAt time.Time) error {
	query := `INSERT INTO tokens (token_id, login, token_hash, expires_at)
	VALUES ($1, $2, $3, $4)`

	_, queryErr := tr.Conn.Exec(query, tokenID, login, tokenHash, expiresAt)
	return queryErr
}

// GetTokenByLogin служит для получения действующего токена пользователя по его логину.
//
// Принимает логин пользователя,
// возвращает токен и возможную ошибку.
func (tr TokenRepository) GetTokenByLogin(login string) (model.RefreshToken, error) {
	query := `SELECT * 
	FROM tokens
	WHERE login=$1 
	AND revoked = false
	ORDER BY expires_at DESC
	LIMIT 1`

	var token model.RefreshToken

	rowErr := tr.Conn.QueryRow(query, login).Scan(&token.TokenID, &token.Login, &token.TokenHash, &token.ExpiresAt, &token.Revoked)
	if rowErr != nil {
		if rowErr == sql.ErrNoRows {
			return model.RefreshToken{}, nil
		} else {
			return model.RefreshToken{}, rowErr
		}
	}

	if token.Revoked {
		return model.RefreshToken{}, nil
	}

	return token, nil
}

// Revoke служит для отзыва всех токенов пользователя.
//
// Принимает логин пользователя,
// возвращает возможную ошибку.
func (tr TokenRepository) Revoke(login string) error {
	query := `UPDATE tokens
	SET revoked = true
	WHERE login = $1`

	_, updateErr := tr.Conn.Exec(query, login)
	return updateErr
}
