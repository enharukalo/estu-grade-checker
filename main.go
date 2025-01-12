package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func connectDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize database schema
	if err := InitDB(ctx, db); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Check for updates immediately at startup
	log.Println("Checking for updates at startup...")
	checkCtx, checkCancel := context.WithTimeout(context.Background(), 30*time.Second)
	CheckForUpdates(checkCtx, bot, db)
	checkCancel()

	// Set up periodic checks
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			checkCtx, checkCancel := context.WithTimeout(context.Background(), 30*time.Second)
			CheckForUpdates(checkCtx, bot, db)
			checkCancel()
		}
	}()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a done channel to signal completion
	done := make(chan bool, 1)

	// Start the bot in a goroutine
	go func() {
		for update := range updates {
			if update.Message != nil {
				HandleMessage(bot, db, update.Message)
			}
		}
		done <- true
	}()

	// Wait for shutdown signal - simplified from select
	<-sigChan
	log.Println("Shutdown signal received")

	// Stop receiving updates
	bot.StopReceivingUpdates()

	// Cancel any ongoing operations
	ticker.Stop()

	// Wait for bot to finish
	<-done

	// Close database connection
	db.Close()
	log.Println("Gracefully shut down")
	os.Exit(0)
}
