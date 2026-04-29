package main

import (
	"fmt"
)

func createTables() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER UNIQUE,
			username TEXT
		)
	`)
	if err != nil {
		fmt.Println("users table error:", err)
		return
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS checklist (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			text TEXT
		)
	`)
	if err != nil {
		fmt.Println("checklist table error:", err)
		return
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("notifications table error:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS limits (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			intent TEXT,
			p2p TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Println("limits table error:", err)
	}
}
