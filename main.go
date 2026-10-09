package main

import (
	//"database/sql"
	"log"
	"net/http"
	//"sync"
	//"encoding/json"

	//"golang.org/x/crypto/bcrypt"
	//"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

func main() {
	createTables() // could be found in dbSceme
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handler)

	log.Println("server have started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}