package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"encoding/json"
	"sync"

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
    old := clients[c.Name]
    clients[c.Name] = c
    mu.Unlock()

    if old != nil && old != c {
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

    cl := &Client{Conn: conn} 

    defer func() {
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

		var req Packet

		err = json.Unmarshal(data, &req)

		if err != nil {
			log.Println("error json unmarshal:", err)
			continue
		}

		switch req.Type {
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
		default:
			log.Println("unknown type:", req.Type)
		}
	}
}
