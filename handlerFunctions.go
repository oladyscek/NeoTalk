package main

import (
	"encoding/json"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// FUNC SEND COULD BE FOUND IN structures.go

func regWS(c *Client, req *Packet) {
	log.Println("REG REQUEST:", req.Name)

	ok := reg(req.Name, req.Password) // reg could be found in dbFunctions.go
	if !ok {
		c.send(Packet{Type: "regStatus", Answer: "failed"})
		return
	}

	token, err := addToken(req.Name) // adding a new token, func also in dbFunctions.go
	if err != nil {
		log.Println("addToken after reg:", err)
		c.send(Packet{Type: "regStatus", Answer: "failed"})
		return
	}

	c.Name = req.Name
	addUserConn(c) // adding Client struct in map after sucsessful reg

	c.send(Packet{Type: "regStatus", Answer: "ok", Token: token})
}

func loginWS(c *Client, req *Packet) {
	log.Println("LOGIN REQUEST:", req.Name)

	ok := login(req.Name, req.Password)
	if !ok {
		c.send(Packet{Type: "loginStatus", Answer: "denied"})
		return
	}

	token, err := addToken(req.Name)
	if err != nil {
		log.Println("addToken:", err)
		c.send(Packet{Type: "loginStatus", Answer: "failed"})
		return
	}

	c.Name = req.Name
	addUserConn(c)

	c.send(Packet{Type: "loginStatus", Answer: "ok", Token: token})
}

func sendMsgWS(c *Client, req *Packet) {
	if c.Name == "" {
		c.send(Packet{Type: "msgStatus", Answer: "not logged in"})
		return
	}

	if c.Name == req.To {
		c.send(Packet{Type: "msgStatus", Answer: "cannot send to yourself"})
		return
	}

	if len(req.Text) == 0 || len(req.Text) > 1000 { // not 0 symbols and not more than 1000
		c.send(Packet{Type: "msgStatus", Answer: "invalidMsgLength"})
		return
	}

	now := time.Now() // not more than 5 messages/second
	if now.Sub(c.LastMsgTime) > time.Second {
		c.MsgCount = 0
		c.LastMsgTime = now
	}
	c.MsgCount++
	if c.MsgCount > 5 {
		c.send(Packet{Type: "msgStatus", Answer: "slow_down"})
		return
	}

	if !req.IsGroup {

		mu.Lock() // finding a recipient
		recipient, _ := clients[req.To]
		mu.Unlock()

		chatID, id, ok := sendMsg(req, c.Name) //send to db, could be found in dbFunctions

		if !ok {
			c.send(Packet{Type: "msgStatus", Answer: "failed"})
			return
		}

		if recipient != nil {
			recipient.send(Packet{Type: "incomedMsg", From: c.Name, To: req.To, MsgId: int(id), Text: req.Text})
		} else {
			go sendPushNotification(req.To, "NeoTalk: "+c.Name, req.Text)
		}

		c.send(Packet{Type: "msgStatus", Answer: "ok", MsgId: int(id), ChatID: int(chatID)})
	} else {
		chatID, id, ok := sendMsg(req, c.Name)

		if !ok {
			c.send(Packet{Type: "msgStatus", Answer: "failed"})
			return
		}

		//searching for all group members
		rows, err := db.Query(`--sql
        SELECT u.name FROM users u
        JOIN chat_members cm ON cm.user_id = u.id
        WHERE cm.chat_id = ? AND u.name != ?
		`, chatID, c.Name)

		if err != nil {
			log.Println("broadcast query error:", err)
			return
		}

		var names []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				log.Println("scan error:", err)
				continue
			}
			names = append(names, name)
		}
		rows.Close()

		mu.Lock()
		members := make([]*Client, 0, len(names))
		offlineNames := make([]string, 0, len(names))
		for _, name := range names {
			if cl, ok := clients[name]; ok {
				members = append(members, cl)
			} else {
				offlineNames = append(offlineNames, name)
			}
		}
		mu.Unlock()

		for _, m := range members { // broadcast for all members
			m.send(Packet{
				Type:   "incomedMsg",
				From:   c.Name,
				MsgId:  int(id),
				ChatID: int(chatID),
				Text:   req.Text,
			})
		}
		for _, name := range offlineNames {
			go sendPushNotification(name, "NeoTalk: "+c.Name, req.Text)
		}
		c.send(Packet{Type: "msgStatus", Answer: "ok", MsgId: int(id), ChatID: int(chatID)})
	}
}

func createChatWS(c *Client, req *Packet) {
	if c.Name == "" {
		c.send(Packet{Type: "newChatStatus", Answer: "not logged in"})
		return
	}

	mu.Lock() // finding a recepient
	r, exist := clients[req.To]
	mu.Unlock()

	ok, id := createChat(req, c.Name) // adding to db, could be found in dbFunctions

	if ok {
		c.send(Packet{Type: "newChatStatus", Answer: "ok", ChatID: id})

		if !req.IsGroup {
			if exist { // sending to recepient alert about new chat
				r.send(Packet{Type: "newChat", ChatID: id})
			}
		} else {
			var users []string // group members except creator
			err := json.Unmarshal(req.Data, &users)

			if err != nil {
				c.send(Packet{Type: "newChatStatus", Answer: "failed"})
				return
			}

			members := make([]*Client, 0, len(users))

			mu.Lock() // getting all online users from group
			for _, user := range users {
				r, exist := clients[user]
				if exist {
					members = append(members, r)
				}
			}
			mu.Unlock()

			for _, r := range members { // sending to all members alert about new chat
				r.send(Packet{Type: "newChat", ChatID: id})
			}
		}

	} else {
		c.send(Packet{Type: "newChatStatus", Answer: "failed"})
	}
}

func getChatsWS(c *Client) {
	if c.Name == "" {
		c.send(Packet{Type: "getChatStatus", Answer: "not logged in"})
		return
	}

	list, ok := getChats(c.Name) // could be found in dbFunctions.go

	if !ok {
		c.send(Packet{Type: "getChatStatus", Answer: "failed"})
		return
	}

	data, err := json.Marshal(list)

	if err != nil {
		log.Println(err)
		c.send(Packet{Type: "getChatStatus", Answer: "failed"})
		return
	}

	c.send(Packet{Type: "getChatStatus", Answer: "ok", Data: data})
}

func getHistoryWS(c *Client, req *Packet) {
	if c.Name == "" {
		c.send(Packet{Type: "getHistoryStatus", Answer: "not logged in"})
		return
	}

	list, ok := getHistory(req, c.Name)

	if !ok {
		c.send(Packet{Type: "getHistoryStatus", Answer: "failed"})
		return
	}

	hasMore := len(list) == 101

	if hasMore {
		list = list[:100]
	}

	data, err := json.Marshal(list)

	if err != nil {
		log.Println(err)
		c.send(Packet{Type: "getHistoryStatus", Answer: "failed"})
		return
	}

	c.send(Packet{Type: "getHistoryStatus", Answer: "ok", Data: data, HasMore: hasMore})
}

func authWS(c *Client, req *Packet) {
	name, ok := authCheck(req.Token) // could be found in dbFunctions.go
	if !ok {
		c.send(Packet{Type: "authStatus", Answer: "denied"})
		return
	}

	ok = updateToken(req.Token) // could be found in dbFunctions.go

	if !ok {
		log.Println("updateToken error")
		c.send(Packet{Type: "authStatus", Answer: "failed"})
		return
	}

	c.Name = name
	addUserConn(c)

	c.send(Packet{Type: "authStatus", Answer: "ok", Name: name})
}

func delMsgWS(c *Client, req *Packet) {
	if c.Name == "" {
		c.send(Packet{Type: "delMsgStatus", Answer: "not logged in"})
		return
	}

	group, isGroup, ok := delMsg(c.Name, req) // could be found in dbFunctions.go

	if !isGroup {
		if ok {
			c.send(Packet{Type: "delMsgStatus", Answer: "ok"})

			mu.Lock()
			r, exist := clients[group[0]] // for DM
			mu.Unlock()

			if exist {
				r.send(Packet{Type: "delMsg", MsgId: req.MsgId})
			}
		} else {
			c.send(Packet{Type: "delMsgStatus", Answer: "failed"})
			return
		}
	} else {
		if ok {
			c.send(Packet{Type: "delMsgStatus", Answer: "ok"})

			members := make([]*Client, 0, len(group))

			mu.Lock()
			for _, u := range group {
				user, exist := clients[u]
				if exist {
					members = append(members, user)
				}
			}
			mu.Unlock()

			for _, cl := range members {
				cl.send(Packet{Type: "delMsg", MsgId: req.MsgId})
			}
		} else {
			c.send(Packet{Type: "delMsgStatus", Answer: "failed"})
			return
		}
	}

}

func chatReadWS(c *Client, req *Packet) {
	if c.Name == "" {
		c.send(Packet{Type: "chatReadStatus", Answer: "not logged in"})
		return
	}

	group, ok := chatRead(c.Name, req)

	if ok {
		c.send(Packet{Type: "chatReadStatus", Answer: "ok"})

		members := make([]*Client, 0, len(group))

		mu.Lock()
		for _, u := range group {
			user, exist := clients[u]
			if exist {
				members = append(members, user)
			}
		}
		mu.Unlock()

		for _, cl := range members {
			cl.send(Packet{Type: "chatRead", ChatID: req.ChatID})
		}
	} else {
		c.send(Packet{Type: "chatReadStatus", Answer: "failed"})
	}
}

func savePushSubWS(c *Client, req *Packet) {
	if c.Name == "" {
		return
	}
	if savePushSub(c.Name, string(req.Data)) {
		log.Println("push subscription saved for:", c.Name)
	}
}
