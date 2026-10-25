package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func createTables() {
	var err error

	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "database.db"
	}

	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"
	db, err = sql.Open("sqlite", dsn)

	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(1)
	
	_, err = db.Exec(`--sql
	CREATE table if not EXISTS users (
		id integer primary key, 
		name text UNIQUE not null,
		password text not null)
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`--sql
	CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	last_used   DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`--sql
	CREATE table if not EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chatname TEXT,                    
		is_group   INTEGER NOT NULL DEFAULT 0,  
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP)
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`--sql
		CREATE TABLE if not EXISTS chat_members (
		chat_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		is_admin bool NOT NULL DEFAULT 0,
		PRIMARY KEY (chat_id, user_id))
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`--sql
		CREATE TABLE if not EXISTS messages (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id    INTEGER NOT NULL,
		sender_id  INTEGER NOT NULL,
		msg        TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP)
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`--sql
	CREATE INDEX IF NOT EXISTS idx_messages_chat_time ON messages(chat_id, created_at DESC)
	`)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("TABLES CREATED")
}
