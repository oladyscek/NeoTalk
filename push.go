package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func sendPushNotification(userName, title, body string) {
	subJSON, ok := getPushSub(userName)
	if !ok {
		return
	}

	sub := &webpush.Subscription{}
	if err := json.Unmarshal([]byte(subJSON), sub); err != nil {
		log.Println("push unmarshal error:", err)
		return
	}

	payload, _ := json.Marshal(map[string]string{"title": title, "body": body})

	authScheme := webpush.Vapid
	if strings.Contains(sub.Endpoint, "fcm.googleapis.com") {
		authScheme = webpush.WebPush
	}

	resp, err := webpush.SendNotification(payload, sub, &webpush.Options{
		Subscriber:      "mailto:silkovmatvey@gmail.com",
		VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		TTL:             30,
		AuthScheme:      authScheme,
	})
	if err != nil {
		log.Println("push send error:", err)
		return
	}
	defer resp.Body.Close()
}
