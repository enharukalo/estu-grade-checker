package main

import "database/sql"

type User struct {
	ID         int
	TelegramID int64
	Cookie     string
	DonemID    string
	Alarm      bool
	Grades     string
}

func GetUserByTelegramID(db *sql.DB, telegramID int64) (*User, error) {
	user := &User{}
	row := db.QueryRow("SELECT id, telegram_id, cookie, donemid, alarm, grades FROM users WHERE telegram_id = ?", telegramID)
	err := row.Scan(&user.ID, &user.TelegramID, &user.Cookie, &user.DonemID, &user.Alarm, &user.Grades)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func InsertUser(db *sql.DB, user *User) error {
	_, err := db.Exec("INSERT INTO users (telegram_id, cookie, donemid, alarm, grades) VALUES (?, ?, ?, ?, ?)",
		user.TelegramID, user.Cookie, user.DonemID, user.Alarm, user.Grades)
	return err
}

func UpdateUser(db *sql.DB, user *User) error {
	_, err := db.Exec("UPDATE users SET cookie = ?, donemid = ?, alarm = ?, grades = ? WHERE telegram_id = ?",
		user.Cookie, user.DonemID, user.Alarm, user.Grades, user.TelegramID)
	return err
}

func UpdateGrades(db *sql.DB, userID int, grades string) error {
	_, err := db.Exec("UPDATE users SET grades = ? WHERE id = ?", grades, userID)
	return err
}