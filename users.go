package main

import (
	"fmt"
)

func saveUser(userID int64, username string) {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO users (telegram_id, username) VALUES (?, ?)
	`, userID, username)
	if err != nil {
		fmt.Println("saveUser error:", err)
	}
}

func getAllUsers() ([]int64, error) {
	rows, err := db.Query(`SELECT telegram_id FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			continue
		}
		users = append(users, userID)
	}
	return users, nil
}

func saveNotification(text string) error {
	_, err := db.Exec(`INSERT INTO notifications (message) VALUES (?)`, text)
	return err
}

func getTodayNotifications() ([]string, error) {
	rows, err := db.Query(`
		SELECT message FROM notifications 
		WHERE date(created_at) = date('now')
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []string
	for rows.Next() {
		var msg string
		rows.Scan(&msg)
		notifications = append(notifications, msg)
	}
	return notifications, nil
}

func updateChecklist(text string) error {
	_, err := db.Exec(`
		INSERT INTO checklist (id, text) VALUES (1, ?)
		ON CONFLICT(id) DO UPDATE SET text = excluded.text
	`, text)
	return err
}
