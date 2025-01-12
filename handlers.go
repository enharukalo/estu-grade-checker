package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if message.IsCommand() {
		switch message.Command() {
		case "start":
			handleStart(bot, message)
		case "cookie":
			handleCookie(ctx, bot, db, message)
		case "donemid":
			handleDonemID(ctx, bot, db, message)
		case "alarm":
			handleAlarm(ctx, bot, db, message)
		case "get":
			handleGet(ctx, bot, db, message)
		}
	}
}

func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome to the ESTU Grade Checker!\nUse /cookie <cookie> to set your cookie.\nUse /donemid <donemid> to set your donemid.\nUse /alarm <true/false> to enable/disable alarms.\nUse /get to get your grades.\nUse /get <course> to get grades for a specific course.\n\nNote: You cannot use /get until both cookie and donemid are set.")
	bot.Send(msg)
}

func handleCookie(ctx context.Context, bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	cookie := strings.TrimSpace(strings.TrimPrefix(message.Text, "/cookie"))
	if cookie == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a valid cookie.")
		bot.Send(msg)
		return
	}

	user, err := GetUserByTelegramID(ctx, db, message.Chat.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Database error: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred. Please try again later.")
			bot.Send(msg)
			return
		}
		user = &User{TelegramID: message.Chat.ID, Cookie: cookie}
		if err := InsertUser(ctx, db, user); err != nil {
			log.Printf("Failed to insert user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to save your cookie. Please try again.")
			bot.Send(msg)
			return
		}
	} else {
		user.Cookie = cookie
		if err := UpdateUser(ctx, db, user); err != nil {
			log.Printf("Failed to update user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to update your cookie. Please try again.")
			bot.Send(msg)
			return
		}
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Cookie updated successfully.")
	bot.Send(msg)
}

func handleDonemID(ctx context.Context, bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	donemID := strings.TrimSpace(strings.TrimPrefix(message.Text, "/donemid"))
	if donemID == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a valid donemid.")
		bot.Send(msg)
		return
	}

	user, err := GetUserByTelegramID(ctx, db, message.Chat.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Database error: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred. Please try again later.")
			bot.Send(msg)
			return
		}
		user = &User{TelegramID: message.Chat.ID, DonemID: donemID}
		if err := InsertUser(ctx, db, user); err != nil {
			log.Printf("Failed to insert user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to save your donemID. Please try again.")
			bot.Send(msg)
			return
		}
	} else {
		user.DonemID = donemID
		if err := UpdateUser(ctx, db, user); err != nil {
			log.Printf("Failed to update user: %v", err)
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to update your donemID. Please try again.")
			bot.Send(msg)
			return
		}
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "DonemID updated successfully.")
	bot.Send(msg)
}

func handleAlarm(ctx context.Context, bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	alarm := strings.TrimSpace(strings.TrimPrefix(message.Text, "/alarm"))
	alarmValue, err := strconv.ParseBool(alarm)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a valid value for alarm (true/false).")
		bot.Send(msg)
		return
	}

	user, err := GetUserByTelegramID(ctx, db, message.Chat.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Please set your cookie and donemid first.")
			bot.Send(msg)
			return
		}
		log.Printf("Database error: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "An error occurred. Please try again later.")
		bot.Send(msg)
		return
	}

	user.Alarm = alarmValue
	if err := UpdateUser(ctx, db, user); err != nil {
		log.Printf("Failed to update user: %v", err)
		msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to update alarm setting. Please try again.")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Alarm preference updated successfully.")
	bot.Send(msg)
}

func handleGet(ctx context.Context, bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	user, err := GetUserByTelegramID(ctx, db, message.Chat.ID)
	if err != nil || user.Cookie == "" || user.DonemID == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You need to set both cookie and donemid before using this command.")
		bot.Send(msg)
		return
	}

	args := strings.TrimSpace(strings.TrimPrefix(message.Text, "/get"))
	if args == "" {
		grades, err := fetchGrades(user.Cookie, user.DonemID)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to fetch grades.")
			bot.Send(msg)
			return
		}

		var oldGrades map[string]map[string]string
		if user.Grades != "" {
			json.Unmarshal([]byte(user.Grades), &oldGrades)
		}

		updates := checkGradeUpdates(oldGrades, grades, user.Alarm)

		userGrades := strings.Builder{}
		for course, gradeDetails := range grades {
			if grade, exists := gradeDetails["Grade"]; exists {
				userGrades.WriteString(fmt.Sprintf("%s: %s\n", course, grade))
			} else {
				userGrades.WriteString(fmt.Sprintf("%s: N/A\n", course))
			}
		}

		updatedGrades, _ := json.Marshal(grades)
		updatedGradesStr := string(updatedGrades)
		if err := UpdateGrades(ctx, db, user.ID, updatedGradesStr); err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to update grades in the database.")
			bot.Send(msg)
			return
		}

		user.Grades = updatedGradesStr

		if len(updates) > 0 && user.Alarm {
			for _, update := range updates {
				msg := tgbotapi.NewMessage(user.TelegramID, update)
				bot.Send(msg)
			}
		}

		msg := tgbotapi.NewMessage(message.Chat.ID, userGrades.String())
		bot.Send(msg)
	} else {
		// Fetch and display all gradetypes and grades for the specified course prefix
		coursePrefix := strings.ToLower(args)

		grades, err := fetchGrades(user.Cookie, user.DonemID)
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to fetch grades.")
			bot.Send(msg)
			return
		}

		matchingCourses := 0
		for course, gradeDetails := range grades {
			if strings.HasPrefix(strings.ToLower(course), coursePrefix) {
				userGrades := strings.Builder{}
				userGrades.WriteString(fmt.Sprintf("Grades for %s:\n", course))
				for gradeType, grade := range gradeDetails {
					userGrades.WriteString(fmt.Sprintf("%s: %s\n", gradeType, grade))
				}
				msg := tgbotapi.NewMessage(message.Chat.ID, userGrades.String())
				bot.Send(msg)
				matchingCourses++
			}
		}

		if matchingCourses == 0 {
			msg := tgbotapi.NewMessage(message.Chat.ID, "No such course found.")
			bot.Send(msg)
		}
	}
}
