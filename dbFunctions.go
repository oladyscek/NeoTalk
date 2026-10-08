package main

import (
	"database/sql"
	"encoding/json"
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

func createChat(req *Packet, creator string) (bool, int64) {
	if !req.IsGroup {

		var isExist int

		err := db.QueryRow(`--sql
			SELECT EXISTS(SELECT 1 from users where name = ?)`, req.To).Scan(&isExist)

		if err != nil {
			log.Println(err)
			return false, -1
		}

		if isExist == 1 {
			var userId int

			err = db.QueryRow(`--sql
				SELECT id from users where name = ?`, req.To).Scan(&userId)

			if err != nil {
				log.Println(err)
				return false, -1
			}

			var creatorId int

			err = db.QueryRow(`--sql
				SELECT id from users where name = ?`, creator).Scan(&creatorId)

			if err != nil {
				log.Println(err)
				return false, -1
			}

			var existingChatID int64

			err = db.QueryRow(`--sql
				SELECT c.id FROM chats c
				JOIN chat_members cm1 ON cm1.chat_id = c.id AND cm1.user_id = ?
				JOIN chat_members cm2 ON cm2.chat_id = c.id AND cm2.user_id = ?
				WHERE c.is_group = 0
				LIMIT 1`, creatorId, userId).Scan(&existingChatID)

			if err == nil {
				return true, existingChatID
			}

			if err != sql.ErrNoRows {
				log.Println("check chat error:", err)
				return false, -1
			}

			res, err := db.Exec(`--sql
				INSERT INTO chats DEFAULT VALUES`)

			if err != nil {
				log.Println(err)
				return false, -1
			}

			id, err := res.LastInsertId()

			if err != nil {
				log.Println(err)
				return false, -1
			}

			_, err = db.Exec(`--sql
				insert into chat_members (chat_id, user_id) values(?,?)`, id, userId)

			if err != nil {
				log.Println(err)
				return false, -1
			}

			_, err = db.Exec(`--sql
				insert into chat_members (chat_id, user_id) values(?,?)`, id, creatorId)

			if err != nil {
				log.Println(err)
				return false, -1
			}
			return true, id
		}
	} else {
		var users []string
		err := json.Unmarshal(req.Data, &users)

		if err != nil {
			log.Println(err)
			return false, -1
		}

		res, err := db.Exec(`--sql
			INSERT INTO chats (chatname, is_group) VALUES (?, 1)`, req.Name)

		if err != nil {
			log.Println(err)
			return false, -1
		}

		id, err := res.LastInsertId()

		if err != nil {
			log.Println(err)
			return false, -1
		}

		for _, user := range users {

			if user == creator {
				continue
			}
			
			var exist int

			err := db.QueryRow(`--sql
				SELECT EXISTS(SELECT 1 from users where name = ?)`, user).Scan(&exist)

			if err != nil {
				log.Println(err)
				return false, -1
			}

			if exist == 1 {

				var userId int

				err = db.QueryRow(`--sql
					SELECT id from users where name = ?`, user).Scan(&userId)

				if err != nil {
					log.Println(err)
					return false, -1
				}

				_, err = db.Exec(`--sql
					insert into chat_members (chat_id, user_id) values(?,?)`, id, userId)

				if err != nil {
					log.Println(err)
					return false, -1
				}
			} else {
				continue
			}
		}
		var creatorId int

		err = db.QueryRow(`--sql
			SELECT id from users where name = ?`, creator).Scan(&creatorId)

		if err != nil {
			log.Println(err)
			return false, -1
		}

		_, err = db.Exec(`--sql
			insert into chat_members (chat_id, user_id, is_admin) values(?,?,?)`, id, creatorId, true)

		if err != nil {
			log.Println(err)
			return false, -1
		}
		return true, id
	}
	return false, -1
}

func sendMsg(req *Packet, sender string) { // todo: write this func
	_, err := db.Exec(`--sql
	`)

	if err != nil {
		log.Println(err)
	}
}
