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

func createChat(req *Packet, creator string) (bool, int) {
	if !req.IsGroup {

		var isExist int

		err := db.QueryRow(`--sql
			SELECT EXISTS(SELECT 1 from users where name = ?)`, req.To).Scan(&isExist) //checking existance of second user

		if err != nil {
			log.Println(err)
			return false, -1
		}

		if isExist == 1 {
			var userId int

			err = db.QueryRow(`--sql
				SELECT id from users where name = ?`, req.To).Scan(&userId) // searching id of second user

			if err != nil {
				log.Println(err)
				return false, -1
			}

			var creatorId int

			err = db.QueryRow(`--sql
				SELECT id from users where name = ?`, creator).Scan(&creatorId) // searching id of creator

			if err != nil {
				log.Println(err)
				return false, -1
			}

			var existingChatID int

			err = db.QueryRow(`--sql
				SELECT c.id FROM chats c
				JOIN chat_members cm1 ON cm1.chat_id = c.id AND cm1.user_id = ?
				JOIN chat_members cm2 ON cm2.chat_id = c.id AND cm2.user_id = ?
				WHERE c.is_group = 0
				LIMIT 1`, creatorId, userId).Scan(&existingChatID) // checking existance of chat between 2 users

			if err == nil {
				return true, existingChatID // if exist just show to user existing chat
			}

			if err != sql.ErrNoRows { // if real error
				log.Println("check chat error:", err)
				return false, -1
			}

			res, err := db.Exec(`--sql
				INSERT INTO chats DEFAULT VALUES`) // creating new chat

			if err != nil {
				log.Println(err)
				return false, -1
			}

			id, err := res.LastInsertId() // id of new chat

			if err != nil {
				log.Println(err)
				return false, -1
			}

			_, err = db.Exec(`--sql
				insert into chat_members (chat_id, user_id) values(?,?)`, id, userId) // adding second user to the chat members

			if err != nil {
				log.Println(err)
				return false, -1
			}

			_, err = db.Exec(`--sql
				insert into chat_members (chat_id, user_id) values(?,?)`, id, creatorId) // adding creator to the chat members

			if err != nil {
				log.Println(err)
				return false, -1
			}
			return true, int(id) // if chat created sucsesfully
		} else {
			return false, -1 // if second user is not exist
		}
	} else { // if group
		var users []string // group members ecxept creator
		err := json.Unmarshal(req.Data, &users)

		if err != nil {
			log.Println(err)
			return false, -1
		}

		res, err := db.Exec(`--sql
			INSERT INTO chats (chatname, is_group) VALUES (?, 1)`, req.Name) // creating a chat where req.Name = name of group

		if err != nil {
			log.Println(err)
			return false, -1
		}

		id, err := res.LastInsertId() // id of new group

		if err != nil {
			log.Println(err)
			return false, -1
		}

		for _, user := range users { // adding all the users to new chat

			if user == creator { // creator shouldnt be added with oter users like that
				continue
			}

			var exist int

			err := db.QueryRow(`--sql
				SELECT EXISTS(SELECT 1 from users where name = ?)`, user).Scan(&exist) // checking existance of each user

			if err != nil {
				log.Println(err)
				return false, -1
			}

			if exist == 1 {

				var userId int

				err = db.QueryRow(`--sql
					SELECT id from users where name = ?`, user).Scan(&userId) // getting id of user to add him to group

				if err != nil {
					log.Println(err)
					return false, -1
				}

				_, err = db.Exec(`--sql
					insert into chat_members (chat_id, user_id) values(?,?)`, id, userId) // adding user to the group

				if err != nil {
					log.Println(err)
					return false, -1
				}
			} else {
				continue // if user is not exist just skip him
			}
		}
		var creatorId int

		err = db.QueryRow(`--sql
			SELECT id from users where name = ?`, creator).Scan(&creatorId) // getting creator id

		if err != nil {
			log.Println(err)
			return false, -1
		}

		_, err = db.Exec(`--sql
			insert into chat_members (chat_id, user_id, is_admin) values(?,?,?)`, id, creatorId, true) // adding creator to the group

		if err != nil {
			log.Println(err)
			return false, -1
		}
		return true, int(id)
	}
}

func sendMsg(req *Packet, sender string) (int64, bool) {
	if !req.IsGroup {
		var isExist int

		err := db.QueryRow(`--sql
			SELECT EXISTS(SELECT 1 from users where name = ?)`, req.To).Scan(&isExist) //checking existance of second user

		if err != nil {
			log.Println(err)
			return -1, false
		}

		if isExist == 1 {
			var userId int

			err = db.QueryRow(`--sql
				SELECT id from users where name = ?`, req.To).Scan(&userId) // searching id of second user

			if err != nil {
				log.Println(err)
				return -1, false
			}

			var senderId int

			err = db.QueryRow(`--sql
				SELECT id from users where name = ?`, sender).Scan(&senderId) // searching id of sender

			if err != nil {
				log.Println(err)
				return -1, false
			}

			var existingChatID int

			err = db.QueryRow(`--sql
				SELECT c.id FROM chats c
				JOIN chat_members cm1 ON cm1.chat_id = c.id AND cm1.user_id = ?
				JOIN chat_members cm2 ON cm2.chat_id = c.id AND cm2.user_id = ?
				WHERE c.is_group = 0
				LIMIT 1`, senderId, userId).Scan(&existingChatID) // checking existance of chat between 2 users

			if err == nil {
				_, err = db.Exec(`--sql
			insert into messages (chat_id, sender_id, msg) values (?,?,?)`, existingChatID, senderId, req.Text)

				if err != nil {
					log.Println(err)
					return -1, false
				}
			} else {
				log.Println(err)
				return -1, false
			}
			return -1, true
		} else { // if second user not exists
			return -1, false
		}
	} else { // if group
		var senderId int

		err := db.QueryRow(`--sql
		select id from users where name = ?`, sender).Scan(&senderId)

		if err != nil {
			log.Println(err)
			return -1, false
		}

		var isExist int

		err = db.QueryRow(`--sql
			SELECT EXISTS(SELECT 1 from chat_members WHERE chat_id = ? AND user_id = ?)`, req.ChatID, senderId).Scan(&isExist)

		if err != nil {
			log.Println(err)
			return -1, false
		}

		if isExist == 1 {
			_, err = db.Exec(`--sql
			insert into messages (chat_id, sender_id, msg) values (?,?,?)`, req.ChatID, senderId, req.Text)

			if err != nil {
				log.Println(err)
				return -1, false
			}

			return int64(req.ChatID), true
		} else {
			return -1, false
		}
	}
}
