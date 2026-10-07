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
		ans, err := json.Marshal(Packet{Type: "regStatus", Answer: "ok"})

		if err != nil {
			log.Println("JSON MARSHAL ERROR:", err)
			return
		}

		conn.WriteMessage(websocket.TextMessage, ans)
	} else {
		ans, err := json.Marshal(Packet{Type: "regStatus", Answer: "username busy"})

		if err != nil {
			log.Println("JSON MARSHAL ERROR:", err)
			return
		}

		conn.WriteMessage(websocket.TextMessage, ans)
	}
}

func loginWS(conn *websocket.Conn, req *Packet) {
	log.Println("LOGIN REQEST:", req.Name)
	ok := login(req.Name, req.Password)

	if !ok {
		data, err := json.Marshal(Packet{Type: "loginStatus", Answer: "denied"})
		if err != nil {
			log.Println("JSON MARSHAL ERROR:", err)
			return
		}
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	mu.Lock()
	_, busy := clients[req.Name]
	mu.Unlock()

	if busy {
		data, err := json.Marshal(Packet{Type: "loginStatus", Answer: "already online"})
		if err != nil {
			log.Println("JSON MARSHAL ERROR:", err)
			return
		}
		conn.WriteMessage(websocket.TextMessage, data)
		return
	}

	addUserConn(conn, req.Name)

	data, err := json.Marshal(Packet{Type: "loginStatus", Answer: "ok"})
	if err != nil {
		log.Println("JSON MARSHAL ERROR:", err)
		return
	}
	conn.WriteMessage(websocket.TextMessage, data)
}


func sendMsgWS(conn *websocket.Conn, req *Packet) { // a lot of bugs here, idk where i am too tierd
	mu.Lock()
	sender := connToName[conn]
	recipientConn, _ := clients[req.To]
	mu.Unlock()

	if sender == "" {
		data, err := json.Marshal(Packet{Type: "msgStatus", Answer: "not logged in"})

		if err != nil {
			log.Println("JSON MARSHAL ERROR:", err)
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, data)

		if err != nil {
			log.Println("send message error:", err)
		}

		return
	}

	sendMsg(req)

	answer, err := json.Marshal(Packet{Type: "incomedMsg", From: sender, To: req.To, Text: req.Text})

	if err != nil {
		log.Println("JSON MARSHAL ERROR:", err)
		return
	}

	err = recipientConn.WriteMessage(websocket.TextMessage, answer)

	if err != nil {
		log.Println("send message error:", err)
		return
	}
}
