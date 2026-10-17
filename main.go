package main

import (
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {
	createTables() // could be found in dbSceme
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/ws", handler)

	go checkTokens()

	log.Println("server have started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}