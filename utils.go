package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type GradeResponse struct {
	OgrenciBirimList []struct {
		TnotlarNotes []struct {
			LabelViewModel struct {
				LanguageMap struct {
					EN string `json:"EN"`
				} `json:"languageMap"`
			} `json:"labelViewModel"`
			NotTreeSet struct {
				Items []struct {
					LabelViewModel struct {
						LanguageMap struct {
							EN string `json:"EN"`
						} `json:"languageMap"`
					} `json:"labelViewModel"`
				} `json:"items"`
			} `json:"notTreeSet"`
		} `json:"tnotlarNotes"`
		SinavTurleriList struct {
			Items []struct {
				LabelViewModel struct {
					LanguageMap struct {
						EN string `json:"EN"`
					} `json:"languageMap"`
				} `json:"labelViewModel"`
			} `json:"items"`
		} `json:"sinavTurleriList"`
	} `json:"ogrenciBirimList"`
}

func CheckForUpdates(ctx context.Context, bot *tgbotapi.BotAPI, db *sql.DB) {
	log.Println("Starting periodic grade check...")
	rows, err := db.QueryContext(ctx, "SELECT id, telegram_id, cookie, donemid, grades, alarm FROM users WHERE alarm = true")
	if err != nil {
		log.Println("Failed to query users:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.TelegramID, &user.Cookie, &user.DonemID, &user.Grades, &user.Alarm); err != nil {
			log.Println("Failed to scan user:", err)
			continue
		}

		log.Printf("Checking grades for user %d...", user.TelegramID)
		newGrades, err := fetchGrades(user.Cookie, user.DonemID)
		if err != nil {
			log.Printf("Failed to fetch grades for user %d: %v\n", user.TelegramID, err)
			msg := tgbotapi.NewMessage(user.TelegramID, "Failed to fetch grades. Please check your cookie and donemid. You can disable alarms using /alarm false.")
			bot.Send(msg)
			continue
		}

		var oldGrades map[string]map[string]string
		if user.Grades != "" {
			if err := json.Unmarshal([]byte(user.Grades), &oldGrades); err != nil {
				log.Printf("Failed to unmarshal old grades for user %d: %v\n", user.TelegramID, err)
				continue
			}
		}

		// Check for updates before updating the database
		updates := checkGradeUpdates(oldGrades, newGrades, user.Alarm)

		// Update database first
		updatedGrades, err := json.Marshal(newGrades)
		if err != nil {
			log.Printf("Failed to marshal new grades for user %d: %v\n", user.TelegramID, err)
			continue
		}

		if err := UpdateGrades(ctx, db, user.ID, string(updatedGrades)); err != nil {
			log.Printf("Failed to update grades in database for user %d: %v\n", user.TelegramID, err)
			continue
		}

		// Send notifications after successful database update
		if len(updates) > 0 {
			log.Printf("Found %d updates for user %d", len(updates), user.TelegramID)
			for _, update := range updates {
				msg := tgbotapi.NewMessage(user.TelegramID, update)
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Failed to send update message to user %d: %v\n", user.TelegramID, err)
				}
			}
		}
	}
	log.Println("Finished periodic grade check")
}

func fetchGrades(cookie string, donemID string) (map[string]map[string]string, error) {
	url := "https://obs.eskisehir.edu.tr/ogrenci/not-gor?donemId=" + donemID

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Cookie", cookie)
	req.Header.Set("Host", "obs.eskisehir.edu.tr")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux x86_64; en-US) Gecko/20100101 Firefox/54.9")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://obs.eskisehir.edu.tr/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var resp *http.Response
	for retries := 3; retries > 0; retries-- {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Non-200 status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch grades. status code: %d", resp.StatusCode)
	}

	var data GradeResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Printf("Failed to decode JSON response: %v", err)
		return nil, err
	}

	log.Printf("Fetched grades data: %+v", data)
	return parseGrades(data), nil
}

func parseGrades(data GradeResponse) map[string]map[string]string {
	grades := make(map[string]map[string]string)

	for _, birim := range data.OgrenciBirimList {
		for _, not := range birim.TnotlarNotes {
			courseName := not.LabelViewModel.LanguageMap.EN
			grades[courseName] = make(map[string]string)

			for i, item := range not.NotTreeSet.Items {
				gradeType := data.OgrenciBirimList[0].SinavTurleriList.Items[i].LabelViewModel.LanguageMap.EN
				grade := item.LabelViewModel.LanguageMap.EN
				if grade != "--" && grade != "" {
					grades[courseName][gradeType] = grade
				}
			}
		}
	}

	log.Printf("Parsed grades: %+v", grades)
	return grades
}

func checkGradeUpdates(oldGrades, newGrades map[string]map[string]string, alarmEnabled bool) []string {
	updates := []string{}
	if len(oldGrades) == 0 || !alarmEnabled {
		return updates
	}

	for course, newCourseGrades := range newGrades {
		oldCourseGrades, exists := oldGrades[course]
		if !exists {
			for gradeType, newGrade := range newCourseGrades {
				updates = append(updates, fmt.Sprintf("The %s for %s has been announced: %s.", gradeType, course, newGrade))
			}
		} else {
			for gradeType, newGrade := range newCourseGrades {
				oldGrade, exists := oldCourseGrades[gradeType]
				if !exists {
					updates = append(updates, fmt.Sprintf("The %s for %s has been announced: %s.", gradeType, course, newGrade))
				} else if newGrade != oldGrade {
					updates = append(updates, fmt.Sprintf("The %s for %s has been updated from %s to %s.", gradeType, course, oldGrade, newGrade))
				}
			}
		}
	}
	return updates
}
