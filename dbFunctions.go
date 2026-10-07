package main

import (
	//"database/sql"
	"log"
	//"net/http"
	//"sync"
	//"encoding/json"

	"golang.org/x/crypto/bcrypt"
	//"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

func reg(name string, password string) bool {
	if name == "" || password == "" {
		return false
	}

	var isExist int

	err := db.QueryRow(`--sql
	SELECT EXISTS(SELECT 1 from users where name = ?)`, name).Scan(&isExist)

	if err != nil {
		log.Println(err)
		return false
	}

	if isExist == 1 {
		return false
	} else {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		_, err = db.Exec(`--sql
		insert into users (name, password) values (?,?)`, name, string(hash))
		if err != nil {
			log.Println("ERROR CANNOT REG:", err)
			return false
		}
		return true
	}
}

func login(name string, password string) bool {
	var hash string

	err := db.QueryRow(`--sql
	SELECT password FROM users WHERE name = ?`, name).Scan(&hash)

	if err != nil {
		log.Println(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == nil {
		return true
	}
	return false
}

func sendMsg(req *Packet) {

}