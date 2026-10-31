package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	createTables() // could be found in dbSceme.go
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/vapid_public", func(w http.ResponseWriter, r *http.Request) { // API endpoint
		w.Write([]byte(os.Getenv("VAPID_PUBLIC_KEY")))
	})
	http.HandleFunc("/ws", handler) // handler could be found in handler.go

	go checkTokens() // deleting old tokens every hour, could be found in background.go

	log.Println("server have started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
