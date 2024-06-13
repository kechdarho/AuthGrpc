package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/domain/models"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)

func (s *Storage) SaveUser(ctx context.Context, email string, login string, phone string, passHash []byte) (int64, error) {
	const op = "storage.sqlite.SaveUser"
	query := "INSERT INTO users(email, login, phone, pass_hash) VALUES ($1, $2, $3, $4)"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, email, login, phone, passHash)

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) User(ctx context.Context, login string) (models.User, error) {
	const op = "storage.sqlite.User"
	query := "SELECT id, login, phone, email, pass_hash, role FROM users WHERE email = $1 OR login =$1 OR phone = $1"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, login)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}
