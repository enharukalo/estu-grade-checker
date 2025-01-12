package main

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID         int
	TelegramID int64
	Cookie     string
	DonemID    string
	Alarm      bool
	Grades     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func InitDB(ctx context.Context, db *sql.DB) error {
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        telegram_id BIGINT UNIQUE NOT NULL,
        cookie TEXT NOT NULL,
        donemid TEXT NOT NULL,
        alarm BOOLEAN NOT NULL DEFAULT TRUE,
        grades JSONB DEFAULT '{}'::jsonb,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );

    CREATE OR REPLACE FUNCTION update_updated_at()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = CURRENT_TIMESTAMP;
        RETURN NEW;
    END;
    $$ language 'plpgsql';

    DROP TRIGGER IF EXISTS update_users_updated_at ON users;
    CREATE TRIGGER update_users_updated_at
        BEFORE UPDATE ON users
        FOR EACH ROW
        EXECUTE FUNCTION update_updated_at();`

	_, err := db.ExecContext(ctx, createTableSQL)
	return err
}

func GetUserByTelegramID(ctx context.Context, db *sql.DB, telegramID int64) (*User, error) {
	user := &User{}
	err := db.QueryRowContext(ctx, `
		SELECT id, telegram_id, cookie, donemid, alarm, grades, created_at, updated_at 
		FROM users WHERE telegram_id = $1`, telegramID).
		Scan(&user.ID, &user.TelegramID, &user.Cookie, &user.DonemID,
			&user.Alarm, &user.Grades, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func InsertUser(ctx context.Context, db *sql.DB, user *User) error {
	if user.Grades == "" {
		user.Grades = "{}"
	}

	return db.QueryRowContext(ctx, `
		INSERT INTO users (telegram_id, cookie, donemid, alarm, grades)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		user.TelegramID, user.Cookie, user.DonemID, user.Alarm, user.Grades).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func UpdateUser(ctx context.Context, db *sql.DB, user *User) error {
	_, err := db.ExecContext(ctx, `
		UPDATE users 
		SET cookie = $1, donemid = $2, alarm = $3, grades = $4
		WHERE telegram_id = $5`,
		user.Cookie, user.DonemID, user.Alarm, user.Grades, user.TelegramID)
	return err
}

func UpdateGrades(ctx context.Context, db *sql.DB, userID int, grades string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE users 
		SET grades = $1
		WHERE id = $2`,
		grades, userID)
	return err
}
