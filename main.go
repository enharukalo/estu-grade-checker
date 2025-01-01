package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB(filepath string) {
	var err error
	db, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        "id" INTEGER PRIMARY KEY AUTOINCREMENT,
        "telegram_id" INTEGER NOT NULL UNIQUE,
        "cookie" TEXT NOT NULL,
        "donemid" TEXT NOT NULL,
        "alarm" BOOLEAN NOT NULL DEFAULT TRUE,
        "grades" TEXT
    );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	InitDB("/app/data/users.db")
	defer db.Close()
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Check for updates immediately at startup
	log.Println("Checking for updates at startup...")
	CheckForUpdates(bot, db)

	// Then set up periodic checks
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			CheckForUpdates(bot, db)
		}
	}()

	for update := range updates {
		if update.Message != nil {
			HandleMessage(bot, db, update.Message)
		}
	}
}
