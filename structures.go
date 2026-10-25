package main

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type HistoryFormat struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
	Time    string `json:"time"`
	Sender  string `json:"sender"`
}

type ChatFormat struct {
	ChatName string `json:"chatName"`
	ID       int    `json:"id"`
	IsGroup  bool   `json:"isGroup"`
}

type Packet struct {
	Type     string          `json:"type"`
	From     string          `json:"from,omitempty"`
	To       string          `json:"to,omitempty"`
	Text     string          `json:"text,omitempty"`
	Name     string          `json:"name,omitempty"`
	Password string          `json:"password,omitempty"`
	Answer   string          `json:"answer,omitempty"`
	ChatID   int             `json:"chatId"`
	MsgId    int             `json:"msgId"`
	IsGroup  bool            `json:"isGroup"`
	Data     json.RawMessage `json:"data,omitempty"`
	Token    string          `json:"token,omitempty"`
	BeforeID int             `json:"beforeId,omitempty"`
	HasMore  bool            `json:"hasMore,omitempty"`
}

type Client struct {
	Conn    *websocket.Conn
	Name    string
	WriteMu sync.Mutex
}

func (c *Client) send(p Packet) {
	data, err := json.Marshal(p)
	if err != nil {
		log.Println("marshal:", err)
		return
	}
	c.WriteMu.Lock()
	defer c.WriteMu.Unlock()
	if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("write:", err)
	}
}
