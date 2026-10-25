package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var validUsername = regexp.MustCompile(`^[a-zA-Z0-9_а-яА-ЯёЁ]{3,20}$`)

func reg(name string, password string) bool {
	if name == "" || password == "" {
		return false
	}

	name = strings.TrimSpace(name)

	if !validUsername.MatchString(name) || password == "" {
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

	name = strings.TrimSpace(name)

	if !validUsername.MatchString(name) || password == "" {
		return false
	}

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
			return int64(existingChatID), true
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

func getChats(sender string) ([]ChatFormat, bool) {
	var userId int

	err := db.QueryRow(`--sql
	select id from users where name = ?`, sender).Scan(&userId)

	if err != nil {
		log.Println(err)
		return nil, false
	}

	data, err := db.Query(`--sql
		SELECT c.id,
		COALESCE(c.chatname, (
			SELECT u.name FROM users u
			JOIN chat_members cm2 ON cm2.user_id = u.id
			WHERE cm2.chat_id = c.id AND u.id != ?
			LIMIT 1
		)) AS name,
		c.is_group
	FROM chats c
	JOIN chat_members cm ON cm.chat_id = c.id
	WHERE cm.user_id = ?`, userId, userId)

	if err != nil {
		log.Println(err)
		return nil, false
	}

	defer data.Close()

	var results []ChatFormat

	for data.Next() {
		var c ChatFormat
		var isGroupInt int
		if err := data.Scan(&c.ID, &c.ChatName, &isGroupInt); err != nil {
			log.Println("scan error:", err)
			continue
		}
		c.IsGroup = isGroupInt == 1
		results = append(results, c)

		if err := data.Err(); err != nil {
			log.Println("rows error:", err)
			return nil, false
		}
	}

	return results, true
}

func getHistory(req *Packet, sender string) ([]HistoryFormat, bool) { // req.ChatID = id of second user or group
	var userId int

	err := db.QueryRow(`--sql
	select id from users where name = ?`, sender).Scan(&userId)

	if err != nil {
		log.Println(err)
		return nil, false
	}

	var isMember int

	err = db.QueryRow(`--sql
    SELECT EXISTS(SELECT 1 FROM chat_members WHERE chat_id = ? AND user_id = ?)
	`, req.ChatID, userId).Scan(&isMember)

	if err != nil {
		log.Println(err)
		return nil, false
	}
	if isMember != 1 {
		log.Println("user not in chat:", req.ChatID)
		return nil, false
	}

	var data *sql.Rows

	if req.BeforeID == 0 {
		data, err = db.Query(`--sql
		SELECT m.id, m.msg, m.created_at, u.name from messages m 
		JOIN users u ON u.id = m.sender_id  
		where m.chat_id = ?
		ORDER BY m.id DESC
		LIMIT 101`, req.ChatID)

		if err != nil {
			log.Println(err)
			return nil, false
		}
	} else {
		data, err = db.Query(`--sql
		SELECT m.id, m.msg, m.created_at, u.name from messages m 
		JOIN users u ON u.id = m.sender_id  
		where m.chat_id = ? and m.id < ?
		ORDER BY m.id DESC
		LIMIT 101`, req.ChatID, req.BeforeID)

		if err != nil {
			log.Println(err)
			return nil, false
		}
	}

	defer data.Close()

	var results []HistoryFormat

	for data.Next() {
		var c HistoryFormat
		if err := data.Scan(&c.Id, &c.Message, &c.Time, &c.Sender); err != nil {
			log.Println("scan error:", err)
			continue
		}
		results = append(results, c)
	}

	if err := data.Err(); err != nil {
		log.Println("rows error:", err)
		return nil, false
	}

	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 { // reverse the slice
		results[i], results[j] = results[j], results[i]
	}

	return results, true
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func addToken(name string) (string, error) {
	var userId int

	err := db.QueryRow(`--sql
	SELECT id FROM users WHERE name = ?`, name).Scan(&userId)

	if err != nil {
		log.Println(err)
		return "", err
	}

	token := generateToken()

	_, err = db.Exec(`--sql
	INSERT INTO sessions (token, user_id) VALUES (?,?)`, token, userId)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return token, nil
}

func authCheck(token string) (string, bool) {
	var name string

	err := db.QueryRow(`--sql
        SELECT u.name FROM sessions s
        JOIN users u ON u.id = s.user_id
        WHERE s.token = ?
    `, token).Scan(&name)

	if err != nil {
		return "", false
	}

	return name, true
}

func updateToken(token string) bool {
	_, err := db.Exec(`--sql
	UPDATE sessions SET last_used = CURRENT_TIMESTAMP WHERE token = ?`, token)

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
