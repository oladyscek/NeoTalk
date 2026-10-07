package main

import (
	"database/sql"
	"log"
	//"net/http"
	//"sync"
	//"encoding/json"

	//"golang.org/x/crypto/bcrypt"
	//"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

func createTables() {
	db, err := sql.Open("sqlite", "database.db")

	if err != nil {
		log.Fatal(err)
	}

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
		PRIMARY KEY (chat_id, user_id))
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`--sql
	CREATE TABLE messages (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id    INTEGER NOT NULL,
    sender_id  INTEGER NOT NULL,
    msg       TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP);

	CREATE INDEX idx_messages_chat_time ON messages(chat_id, created_at DESC)
	`)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`--sql
	CREATE INDEX idx_messages_chat_time ON messages(chat_id, created_at DESC)
	`)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("TABLES CREATED")
}