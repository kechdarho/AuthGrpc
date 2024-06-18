package sqlite

import (
	"AuthGrpc/internal/domain/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func (s *Storage) SaveUser(ctx context.Context, email string, login string, phone string, passHash []byte) error {
	const op = "storage.sqlite.SaveUser"
	query := "INSERT INTO users(email, login, phone, pass_hash) VALUES (?, ?, ?, ?)"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)
	_, err = stmt.ExecContext(ctx, email, login, phone, passHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetUser(ctx context.Context, login string) (models.User, error) {
	const op = "storage.sqlite.GetUser"
	query := "SELECT id, login, phone, email, pass_hash, role FROM users WHERE email = ? OR login = ? OR phone = ?"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	row := stmt.QueryRowContext(ctx, login, login, login)

	var user models.User
	err = row.Scan(&user.ID, &user.Login, &user.Phone, &user.Email, &user.PassHash, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) UpdateUser(ctx context.Context, id int64, updates map[string]interface{}) (bool, error) {
	const op = "storage.sqlite.UpdateUser"
	query := "UPDATE users SET "
	var args []interface{}

	argCounter := 1
	validColumns := map[string]bool{
		"pass_hash": true,
		"email":     true,
		"phone":     true,
	}

	for column, value := range updates {
		if !validColumns[column] {
			return false, fmt.Errorf("%s: invalid column name %s", op, column)
		}
		if isEmpty(value) {
			continue
		}
		if argCounter > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", column, argCounter)
		args = append(args, value)
		argCounter++
	}

	if argCounter == 1 {
		return false, fmt.Errorf("%s: no valid fields to update", op)
	}

	query += fmt.Sprintf(" WHERE id = $%d", argCounter)
	args = append(args, id)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	_, err = stmt.ExecContext(ctx, args...)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return true, nil
}

func (s *Storage) SaveToken(ctx context.Context, id int64, token string, expiresAt time.Time) error {
	const op = "storage.sqlite.SaveToken"
	query := "INSERT INTO resetTokens (user_id, token, expires_at) VALUES (?, ?, ?)"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	_, err = stmt.ExecContext(ctx, id, token, expiresAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) GetToken(ctx context.Context, token string) (models.ResetToken, error) {
	const op = "storage.sqlite.GetToken"
	query := "SELECT id, user_id, token, expires_at FROM resetTokens WHERE token = ?"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return models.ResetToken{}, fmt.Errorf("%s: %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	row := stmt.QueryRowContext(ctx, token)

	var resetToken models.ResetToken
	err = row.Scan(&resetToken.ID, &resetToken.UserID, &resetToken.Token, &resetToken.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ResetToken{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return models.ResetToken{}, fmt.Errorf("%s: %w", op, err)
	}
	return resetToken, nil
}

func (s *Storage) DeleteToken(ctx context.Context, id int64) error {
	const op = "storage.sqlite.DeleteToken"
	query := "DELETE FROM resetTokens WHERE id = ?"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func isEmpty(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case []byte:
		return len(v) == 0
	default:
		return false
	}
}
