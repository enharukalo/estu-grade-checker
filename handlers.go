package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleMessage(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			handleStart(bot, message)
		case "cookie":
			handleCookie(bot, db, message)
		case "donemid":
			handleDonemID(bot, db, message)
		case "alarm":
			handleAlarm(bot, db, message)
		case "get":
			handleGet(bot, db, message)
		}
	}
}

func handleStart(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Welcome to the ESTU Grade Checker!\nUse /cookie <cookie> to set your cookie.\nUse /donemid <donemid> to set your donemid.\nUse /alarm <true/false> to enable/disable alarms.\nUse /get to get your grades.\nUse /get <course> to get grades for a specific course.\n\nNote: You cannot use /get until both cookie and donemid are set.")
	bot.Send(msg)
}

func handleCookie(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	cookie := strings.TrimSpace(strings.TrimPrefix(message.Text, "/cookie"))
	if cookie == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a valid cookie.")
		bot.Send(msg)
		return
	}

	user, err := GetUserByTelegramID(db, message.Chat.ID)
	if err != nil {
		user = &User{TelegramID: message.Chat.ID, Cookie: cookie}
		InsertUser(db, user)
	} else {
		user.Cookie = cookie
		UpdateUser(db, user)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "Cookie updated successfully.")
	bot.Send(msg)
}

func handleDonemID(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	donemID := strings.TrimSpace(strings.TrimPrefix(message.Text, "/donemid"))
	if donemID == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a valid donemid.")
		bot.Send(msg)
		return
	}

	user, err := GetUserByTelegramID(db, message.Chat.ID)
	if err != nil {
		user = &User{TelegramID: message.Chat.ID, DonemID: donemID}
		InsertUser(db, user)
	} else {
		user.DonemID = donemID
		UpdateUser(db, user)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, "DonemID updated successfully.")
	bot.Send(msg)
}

func handleAlarm(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	alarm := strings.TrimSpace(strings.TrimPrefix(message.Text, "/alarm"))
	alarmValue, err := strconv.ParseBool(alarm)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please provide a valid value for alarm (true/false).")
		bot.Send(msg)
		return
	}

	user, err := GetUserByTelegramID(db, message.Chat.ID)
	if err != nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please set your cookie and donemid first.")
		bot.Send(msg)
		return
	}

	user.Alarm = alarmValue
	UpdateUser(db, user)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Alarm preference updated successfully.")
	bot.Send(msg)
}

func handleGet(bot *tgbotapi.BotAPI, db *sql.DB, message *tgbotapi.Message) {
	user, err := GetUserByTelegramID(db, message.Chat.ID)
	if err != nil || user.Cookie == "" || user.DonemID == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "You need to set both cookie and donemid before using this command.")
		bot.Send(msg)
		return
	}

	args := strings.TrimSpace(strings.TrimPrefix(message.Text, "/get"))
	if args == "" {
		// Fetch and display "Grade" gradetype for all courses
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

		userGrades := strings.Builder{}
		for course, gradeDetails := range grades {
			if grade, exists := gradeDetails["Grade"]; exists {
				userGrades.WriteString(fmt.Sprintf("%s: %s\n", course, grade))
			} else {
				userGrades.WriteString(fmt.Sprintf("%s: N/A\n", course))
			}
		}

		updates := checkGradeUpdates(oldGrades, grades, user.Alarm)
		if user.Grades != "" && len(updates) > 0 && user.Alarm {
			for _, update := range updates {
				msg := tgbotapi.NewMessage(user.TelegramID, update)
				bot.Send(msg)
			}
		}

		updatedGrades, _ := json.Marshal(grades)
		err = UpdateGrades(db, user.ID, string(updatedGrades))
		if err != nil {
			msg := tgbotapi.NewMessage(message.Chat.ID, "Failed to update grades in the database.")
			bot.Send(msg)
			return
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