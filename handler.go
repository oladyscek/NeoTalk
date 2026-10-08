package main

import (
	"database/sql"
	"log"
	"net/http"

	"encoding/json"
	"sync"

	//"golang.org/x/crypto/bcrypt"
	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

type Packet struct {
	Type     string          `json:"type"`
	From     string          `json:"from,omitempty"`
	To       string          `json:"to,omitempty"`
	Text     string          `json:"text,omitempty"`
	Name     string          `json:"name,omitempty"`
	Password string          `json:"password,omitempty"`
	Answer   string          `json:"answer,omitempty"`
	ChatID   int             `json:"chatId"`
	IsGroup  bool            `json:"IsGroup"`
	Data     json.RawMessage `json:"data,omitempty"`
}

var db *sql.DB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	mu         sync.Mutex
	clients    = make(map[string]*websocket.Conn)
	connToName = make(map[*websocket.Conn]string)
)

func addUserConn(conn *websocket.Conn, name string) {
	mu.Lock()
	clients[name] = conn
	connToName[conn] = name
	mu.Unlock()
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("upgrade error", err)
		return
	}

	defer conn.Close()

	defer func() {
		mu.Lock()
		name := connToName[conn]
		if name != "" {
			delete(clients, name)
			delete(connToName, conn)
			log.Println("отключился:", name)
		}
		mu.Unlock()
	}()

	log.Println("new connection", conn.RemoteAddr())

	for {
		_, data, err := conn.ReadMessage()

		if err != nil {
			log.Println("error reading message:", err)
			return
		}

		var req Packet

		err = json.Unmarshal(data, &req)

		if err != nil {
			log.Println("error json unmarshal:", err)
			continue
		}

		switch req.Type {
		case "reg":
			regWS(conn, &req)
		case "login":
			loginWS(conn, &req)
		case "sendMsg":
			sendMsgWS(conn, &req)
		case "getHistory":
			continue
		case "createChat":
			createChatWS(conn, &req)
		case "getChats":
			continue
		}
	}
}
