package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func checkTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		res, err := db.Exec(`--sql
            DELETE FROM sessions
            WHERE last_used < datetime('now', '-20 days')
        `)
		if err != nil {
			log.Println("session cleanup error:", err)
			continue
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			log.Printf("cleaned up %d expired sessions", n)
		}
	}
}

func heartBeat(cl *Client) {
	ticker := time.NewTicker(25 * time.Second) // ping every 25 sec
	defer ticker.Stop()
	for range ticker.C {
		cl.WriteMu.Lock()
		err := cl.Conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
		cl.WriteMu.Unlock()

		if err != nil {
			return
		}
	}
}
