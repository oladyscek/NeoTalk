package main

import (
    "log"
    "time"
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