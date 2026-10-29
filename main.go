package main

import (
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {
	createTables() // could be found in dbSceme.go
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handler) // handler could be found in handler.go

	go checkTokens() // deleting old tokens eveey hour, could be found in background.go

	log.Println("server have started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}