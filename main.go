package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func serveWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("user")
	if username == "" {
		http.Error(w, "username required", http.StatusBadRequest)
		return
	}

	// Capture FCM token if provided
	fcmToken := r.URL.Query().Get("fcm_token")
	if fcmToken != "" {
		hub.SetFCMToken(username, fcmToken)
		log.Printf("Stored FCM token for user: %s", username)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		username: username,
		conn:     conn,
		send:     make(chan Message, 10),
	}

	hub.register <- client

	go client.writeLoop()
	go client.readLoop(hub)
}

func main() {
	// Initialize Firebase Admin SDK
	if err := InitFirebase("./serviceAccountKey.json"); err != nil {
		log.Printf("Warning: Firebase initialization failed: %v", err)
		log.Println("Server will run without FCM support")
	}

	hub := NewHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	log.Println("Signaling server running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
