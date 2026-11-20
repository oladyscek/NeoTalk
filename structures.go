package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// HistoryFormat represents a single message in the chat history
type HistoryFormat struct {
	// --- Message Identity ---
	Id     int    `json:"id"`
	Sender string `json:"sender"` // Original username
	Time   string `json:"time"`

	// --- Content ---
	Message       string `json:"message"`
	IsEdited      bool   `json:"isEdited"`
	ForwardedFrom string `json:"forwardedFrom"`

	// --- Status ---
	IsRead bool `json:"isRead"`
}

// ChatFormat represents a chat entry in the sidebar list
type ChatFormat struct {
	// --- Chat Identity & Display ---
	ID       int    `json:"id"`
	IsGroup  bool   `json:"isGroup"`
	ChatName string `json:"chatName"` // Display name (with custom nickname applied)
	PeerName string `json:"peerName"` // Original username (used for routing in DMs)

	// --- Last Message Preview ---
	LastMsg       string `json:"lastMsg"`
	LastMsgSender string `json:"lastMsgSender"`
	LastMsgIsRead bool   `json:"lastMsgIsRead"`
	LastUsed      string `json:"lastUsed"`

	// --- Status ---
	Unread int `json:"unread"`
}

type Packet struct {
	// --- Meta & Status ---
	Type   string `json:"type"`
	Answer string `json:"answer,omitempty"` // ok, failed, denied, not logged in

	// --- Auth & Identity ---
	Token    string `json:"token,omitempty"`
	Name     string `json:"name,omitempty"` // username, group name, or custom name
	Password string `json:"password,omitempty"`

	// --- Routing & Chat Context ---
	From    string `json:"from,omitempty"` // sender (server -> client)
	To      string `json:"to,omitempty"`   // recipient (client -> server for DM)
	ChatID  int    `json:"chatId"`
	IsGroup bool   `json:"isGroup"`

	// --- Message Payload ---
	MsgId         int    `json:"msgId"`
	Text          string `json:"text,omitempty"`
	Time          string `json:"time,omitempty"`
	ForwardMsgId  int    `json:"forwardMsgId,omitempty"`
	ForwardedFrom string `json:"forwardedFrom,omitempty"`

	// --- Pagination (History) ---
	BeforeID int  `json:"beforeId,omitempty"`
	HasMore  bool `json:"hasMore,omitempty"`

	// --- Arbitrary Data ---
	Data json.RawMessage `json:"data,omitempty"` // arrays (user lists, chats) or raw objects
}

type Client struct {
	Conn    *websocket.Conn
	Name    string
	WriteMu sync.Mutex

	MsgCount    int
	LastMsgTime time.Time
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
