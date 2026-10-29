package main

import (
	"encoding/json"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func regWS(c *Client, req *Packet) {
	log.Println("REG REQUEST:", req.Name)

	ok := reg(req.Name, req.Password)
	if !ok {
		c.send(Packet{Type: "regStatus", Answer: "failed"})
		return
	}

	token, err := addToken(req.Name)
	if err != nil {
		log.Println("addToken after reg:", err)
		c.send(Packet{Type: "regStatus", Answer: "failed"})
		return
	}

	c.Name = req.Name
	addUserConn(c)

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
	if len(req.Text) == 0 || len(req.Text) > 1000 {
		c.send(Packet{Type: "msgStatus", Answer: "invalidMsgLength"})
		return
	}

	now := time.Now()
	if now.Sub(c.LastMsgTime) > time.Second {
		c.MsgCount = 0
		c.LastMsgTime = now
	}
	c.MsgCount++
	if c.MsgCount > 5 {
		c.send(Packet{Type: "msgStatus", Answer: "slow_down"})
		return
	}

	mu.Lock()
	recipient, _ := clients[req.To]
	mu.Unlock()

	if c.Name == "" {
		c.send(Packet{Type: "msgStatus", Answer: "not logged in"})
		return
	}

	if c.Name == req.To {
		c.send(Packet{Type: "msgStatus", Answer: "cannot send to yourself"})
		return
	}

	if !req.IsGroup {

		chatID, id, ok := sendMsg(req, c.Name) //send to db, could be found in dbFunctions

		if !ok {
			c.send(Packet{Type: "msgStatus", Answer: "failed"})
			return
		}

		if recipient != nil {
			recipient.send(Packet{Type: "incomedMsg", From: c.Name, To: req.To, MsgId: int(id), Text: req.Text})
		}
		c.send(Packet{Type: "msgStatus", Answer: "ok", MsgId: int(id), ChatID: int(chatID)})
	} else {
		chatID, id, ok := sendMsg(req, c.Name)

		if !ok {
			c.send(Packet{Type: "msgStatus", Answer: "failed"})
			return
		}

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
		for _, name := range names {
			if cl, ok := clients[name]; ok {
				members = append(members, cl)
			}
		}
		mu.Unlock()

		for _, m := range members {
			m.send(Packet{
				Type:   "incomedMsg",
				From:   c.Name,
				MsgId:  int(id),
				ChatID: int(chatID),
				Text:   req.Text,
			})
		}
		c.send(Packet{Type: "msgStatus", Answer: "ok", MsgId: int(id), ChatID: int(chatID)})
	}
}

func createChatWS(c *Client, req *Packet) {
	mu.Lock()
	r, exist := clients[req.To]
	mu.Unlock()

	if c.Name == "" {
		c.send(Packet{Type: "newChatStatus", Answer: "not logged in"})
		return
	}

	ok, id := createChat(req, c.Name) // adding to db, could be found in dbFunctions

	if ok {
		c.send(Packet{Type: "newChatStatus", Answer: "ok", ChatID: id})

		if !req.IsGroup {
			if exist {
				r.send(Packet{Type: "newChat", ChatID: id})
			}
		} else {
			var users []string // group members except creator
			err := json.Unmarshal(req.Data, &users)

			if err != nil {
				c.send(Packet{Type: "newChatStatus", Answer: "failed"})
				return
			}
			for _, user := range users {
				mu.Lock()
				r, exist := clients[user]
				mu.Unlock()

				if exist {
					r.send(Packet{Type: "newChat", ChatID: id})
				}
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

	list, ok := getChats(c.Name)

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
	name, ok := authCheck(req.Token)
	if !ok {
		c.send(Packet{Type: "authStatus", Answer: "denied"})
		return
	}

	ok = updateToken(req.Token)

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

	group, isGroup, ok := delMsg(c.Name, req)

	if !isGroup {
		if ok {
			c.send(Packet{Type: "delMsgStatus", Answer: "ok"})

			mu.Lock()
			r, exist := clients[group[0]]
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
