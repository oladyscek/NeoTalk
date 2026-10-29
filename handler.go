package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

var db *sql.DB

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || strings.Contains(origin, r.Host)
	},
}

var (
	mu      sync.Mutex
	clients = make(map[string]*Client) // name -> client
)

func addUserConn(c *Client) {
	mu.Lock()
	old := clients[c.Name] // looking for old session
	clients[c.Name] = c // adding new session
	mu.Unlock()

	if old != nil && old != c {
		msg := websocket.FormatCloseMessage(4000, "kicked") // 4000 - accouns is online frov another device
		old.Conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second)) // give a second to packet to be deilvered to old session
		old.Conn.Close()
		log.Println("kicked old session for:", c.Name)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade error", err)
		return
	}

	conn.SetReadLimit(64 * 1024) // one JSON is 64 KByte maximum

	cl := &Client{Conn: conn} // adding new client without name yet 
	// struct Client could be found in structures.go

	defer func() { // if disconnected
		conn.Close() 
		mu.Lock()
		if cl.Name != "" && clients[cl.Name] == cl {
			delete(clients, cl.Name)
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

		var req Packet // could be found in structures.go

		err = json.Unmarshal(data, &req)

		if err != nil {
			log.Println("error json unmarshal:", err)
			continue
		}

		switch req.Type { // all WS functions could be found in handlerFunctions.go
		case "reg":
			regWS(cl, &req)
		case "login":
			loginWS(cl, &req)
		case "sendMsg":
			sendMsgWS(cl, &req)
		case "getHistory":
			getHistoryWS(cl, &req)
		case "createChat":
			createChatWS(cl, &req)
		case "getChats":
			getChatsWS(cl)
		case "auth":
			authWS(cl, &req)
		case "delMsg":
			delMsgWS(cl, &req)
		case "chatRead":
			chatReadWS(cl, &req)
		default:
			log.Println("unknown type:", req.Type)
		}
	}
}
