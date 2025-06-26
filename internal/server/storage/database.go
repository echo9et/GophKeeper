package storage

import (
	"context"
	"database/sql"
	"fmt"

	"log/slog"

	"GophKeeper.ru/internal/entities"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Database — структура, представляющая подключение к базе данных.
type Database struct {
	addr string
	conn *sql.DB
}

// New создаёт новый экземпляр Database и открывает соединение.
func New(addr string) (*Database, error) {
	db := &Database{addr: addr}
	if err := db.Open(addr); err != nil {
		return nil, err
	}
	return db, nil
}

// Open устанавливает соединение с PostgreSQL через драйвер pgx.
func (db *Database) Open(addr string) error {
	conn, err := sql.Open("pgx", addr)
	if err != nil {
		slog.Error("Failed to open database connection",
			"error", err,
			"method", "Open")
		return err
	}

	db.conn = conn
	return nil
}

// CompareUser заглушка для примера. Можно реализовать проверку хэша.
func (db *Database) CompareUser(hash string) error {
	return nil
}

// Ping проверяет активность соединения с БД.
func (db *Database) Ping() error {
	if err := db.conn.Ping(); err != nil {
		slog.Warn("Database ping failed",
			"error", err,
			"method", "Ping")
		return err
	}
	return nil
}

// Stop закрывает соединение с БД и логирует событие.
func (db *Database) Stop() error {
	slog.Info("Database is stopping",
		"method", "Stop")
	return nil
}

// User возвращает пользователя по его логину.
func (db *Database) User(ctx context.Context, login string) (*entities.User, error) {
	var user entities.User
	err := db.conn.QueryRowContext(ctx,
		"SELECT id, name, password FROM users WHERE name = $1", login).
		Scan(&user.ID, &user.Login, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("Failed to fetch user by login",
			"login", login,
			"error", err,
			"method", "User")
		return nil, err
	}

	return &user, nil
}

// UserFromID возвращает пользователя по его ID.
func (db *Database) UserFromID(ctx context.Context, id int) (*entities.User, error) {
	var user entities.User
	err := db.conn.QueryRowContext(ctx,
		"SELECT id, name, password, is_disable FROM users WHERE id = $1", id).
		Scan(&user.ID, &user.Login, &user.Password, &user.IsDisable)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		slog.Error("Failed to fetch user by ID",
			"user_id", id,
			"error", err,
			"method", "UserFromID")
		return nil, err
	}

	return &user, nil
}

// GetData возвращает все данные пользователя по его ID.
func (db *Database) GetData(ctx context.Context, id int) (*entities.Update, error) {
	out := entities.NewUpdate()
	rows, err := db.conn.QueryContext(ctx, "SELECT name, value FROM data WHERE user_id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return out, nil
		}
		slog.Error("Failed to query user data",
			"user_id", id,
			"error", err,
			"method", "GetData")
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r entities.Record
		err := rows.Scan(&r.Key, &r.Value)
		if err != nil {
			slog.Error("Failed to scan row in GetData",
				"error", err,
				"method", "GetData")
			return nil, err
		}
		out.Data = append(out.Data, r)
	}

	return out, nil
}

// GetCountUpdate возвращает текущий номер обновления пользователя.
func (db *Database) GetCountUpdate(ctx context.Context, userID int) (int, error) {
	var countUpdate int

	err := db.conn.QueryRowContext(ctx,
		"SELECT update_id FROM users WHERE id = $1", userID).
		Scan(&countUpdate)
	if err != nil {
		slog.Error("Failed to get update_id",
			"user_id", userID,
			"error", err,
			"method", "GetCountUpdate")
		return countUpdate, err
	}

	return countUpdate, nil
}

// UpdateRecord добавляет или обновляет запись пользователя в БД.
func (db *Database) UpdateRecord(ctx context.Context, userID int, r entities.Record) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Failed to begin transaction",
			"error", err,
			"method", "UpdateRecord")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
        INSERT INTO data (user_id, name, value)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, name) DO UPDATE SET
            value = EXCLUDED.value,
            updated_at = CURRENT_TIMESTAMP
    `, userID, r.Key, r.Value)

	if err != nil {
		slog.Error("Failed to update record",
			"user_id", userID,
			"key", r.Key,
			"error", err,
			"method", "UpdateRecord")
		return fmt.Errorf("failed to update: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE users
        SET update_id = update_id + 1
        WHERE id = $1
    `, userID)

	if err != nil {
		slog.Error("Failed to update user's update_id",
			"user_id", userID,
			"error", err,
			"method", "UpdateRecord")
		return fmt.Errorf("failed to update user's update_id: %w", err)
	}

	if err = tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction",
			"error", err,
			"method", "UpdateRecord")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RemoveRecord удаляет запись пользователя из БД.
func (db *Database) RemoveRecord(ctx context.Context, userID int, key string) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("Failed to begin transaction",
			"error", err,
			"method", "RemoveRecord")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
        DELETE FROM data
        WHERE user_id = $1 AND name = $2
    `, userID, key)

	if err != nil {
		slog.Error("Failed to delete record",
			"user_id", userID,
			"key", key,
			"error", err,
			"method", "RemoveRecord")
		return fmt.Errorf("failed to delete record: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
        UPDATE users
        SET update_id = update_id + 1
        WHERE id = $1
    `, userID)

	if err != nil {
		slog.Error("Failed to update user's update_id after deletion",
			"user_id", userID,
			"error", err,
			"method", "RemoveRecord")
		return fmt.Errorf("failed to update user's update_id: %w", err)
	}

	if err = tx.Commit(); err != nil {
		slog.Error("Failed to commit transaction",
			"error", err,
			"method", "RemoveRecord")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
