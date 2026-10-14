package main

import (
	//"database/sql"
	"log"
	//"net/http"
	//"sync"
	"encoding/json"

	//"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

func regWS(conn *websocket.Conn, req *Packet) {
	log.Println("REG REQEST:", req.Name)
	ok := reg(req.Name, req.Password)

	if ok {
		sendPacket(conn, Packet{Type: "regStatus", Answer: "ok"})
	} else {
		sendPacket(conn, Packet{Type: "regStatus", Answer: "username busy"})
	}
}

func loginWS(conn *websocket.Conn, req *Packet) {
	log.Println("LOGIN REQEST:", req.Name)
	ok := login(req.Name, req.Password)

	if !ok {
		sendPacket(conn, Packet{Type: "loginStatus", Answer: "denied"})
		return
	}

	mu.Lock()
	_, busy := clients[req.Name]
	mu.Unlock()

	if busy {
		sendPacket(conn, Packet{Type: "loginStatus", Answer: "already online"})
		return
	}

	addUserConn(conn, req.Name)

	sendPacket(conn, Packet{Type: "loginStatus", Answer: "ok"})
}

func sendPacket(conn *websocket.Conn, req Packet) {
	answer, err := json.Marshal(req)

	if err != nil {
		log.Println("JSON MARSHAL ERROR:", err)
		return
	}

	err = conn.WriteMessage(websocket.TextMessage, answer)

	if err != nil {
		log.Println("send message error:", err)
		return
	}
}

func sendMsgWS(conn *websocket.Conn, req *Packet) {
	mu.Lock()
	sender := connToName[conn]
	recipientConn, _ := clients[req.To]
	mu.Unlock()

	if sender == "" {
		sendPacket(conn, Packet{Type: "msgStatus", Answer: "not logged in"})
		return
	}

	if sender == req.To {
		sendPacket(conn, Packet{Type: "msgStatus", Answer: "cannot send to yourself"})
		return
	}

	if !req.IsGroup {

		_, ok := sendMsg(req, sender) //send to db, could be found in dbFunctions

		if !ok {
			sendPacket(conn, Packet{Type: "msgStatus", Answer: "failed"})
			return
		}

		if recipientConn != nil {
			sendPacket(recipientConn, Packet{Type: "incomedMsg", From: sender, To: req.To, Text: req.Text})
		}
		sendPacket(conn, Packet{Type: "msgStatus", Answer: "ok"})
	} else {
		chatID, ok := sendMsg(req, sender)

		if !ok {
			sendPacket(conn, Packet{Type: "msgStatus", Answer: "failed"})
			return
		}

		rows, err := db.Query(`--sql
        SELECT u.name FROM users u
        JOIN chat_members cm ON cm.user_id = u.id
        WHERE cm.chat_id = ? AND u.name != ?
		`, chatID, sender)
		if err != nil {
			log.Println("broadcast query error:", err)
			return
		}
		defer rows.Close()

		mu.Lock()
		defer mu.Unlock()

		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				log.Println("scan error:", err)
				continue
			}

			if recipientConn, ok := clients[name]; ok { // if user online send by websocket
				sendPacket(recipientConn, Packet{
					Type:   "incomedMsg",
					From:   sender,
					ChatID: int(chatID),
					Text:   req.Text,
				})
			}
		}
	}
}

func createChatWS(conn *websocket.Conn, req *Packet) {
	mu.Lock()
	creator := connToName[conn]
	mu.Unlock()

	if creator == "" {
		sendPacket(conn, Packet{Type: "msgStatus", Answer: "not logged in"})
		return
	}

	ok, id := createChat(req, creator) // adding to db, could be found in dbFunctions

	if ok {
		sendPacket(conn, Packet{Type: "newChatStatus", Answer: "ok", ChatID: id})
	} else {
		sendPacket(conn, Packet{Type: "newChatStatus", Answer: "failed"})
	}
}

func getChatsWS(conn *websocket.Conn) {
	mu.Lock()
	sender, ok := connToName[conn]
	mu.Unlock()

	if !ok {
		sendPacket(conn, Packet{Type: "getChatStatus", Answer: "not logged in"})
		return
	}

	list, ok := getChats(sender)

	if !ok {
		sendPacket(conn, Packet{Type: "getChatStatus", Answer: "failed"})
		return
	}

	data, err := json.Marshal(list)

	if err != nil {
		log.Println(err)
		sendPacket(conn, Packet{Type: "getChatStatus", Answer: "failed"})
		return
	}

	sendPacket(conn, Packet{Type: "getChatStatus", Answer: "ok", Data: data})
}

func getHistoryWS(conn *websocket.Conn, req *Packet) {
	mu.Lock()
	sender, ok := connToName[conn]
	mu.Unlock()

	if !ok {
		sendPacket(conn, Packet{Type: "getHistoryStatus", Answer: "not logged in"})
		return
	}

	list, ok := getHistory(req, sender)

	if !ok {
		sendPacket(conn, Packet{Type: "getHistoryStatus", Answer: "failed"})
		return
	}

	data, err := json.Marshal(list)

	if err != nil {
		log.Println(err)
		sendPacket(conn, Packet{Type: "getHistoryStatus", Answer: "failed"})
		return
	}

	sendPacket(conn, Packet{Type: "getHistoryStatus", Answer: "ok", Data: data})
}
