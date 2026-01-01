package main

import "sync"

type Hub struct {
	clients    map[string]*Client
	activeCall map[string]string // user -> callID
	fcmTokens  map[string]string // user -> FCM token (persistent)
	mu         sync.RWMutex      // protects fcmTokens
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		activeCall: make(map[string]string),
		fcmTokens:  make(map[string]string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// SetFCMToken stores the FCM token for a user (thread-safe)
func (h *Hub) SetFCMToken(username, token string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.fcmTokens[username] = token
}

// GetFCMToken retrieves the FCM token for a user (thread-safe)
func (h *Hub) GetFCMToken(username string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	token, ok := h.fcmTokens[username]
	return token, ok
}

func (h *Hub) run() {
	for {
		select {

		case c := <-h.register:
			h.clients[c.username] = c
			h.broadcastPresence(c.username, true)

		case c := <-h.unregister:
			delete(h.clients, c.username)
			delete(h.activeCall, c.username)
			h.broadcastPresence(c.username, false)
		}
	}
}

func (h *Hub) broadcastPresence(user string, online bool) {
	for _, c := range h.clients {
		c.send <- Message{
			Type: "presence",
			Data: map[string]interface{}{
				"user":   user,
				"online": online,
			},
		}
	}
}

func (h *Hub) handleMessage(sender *Client, msg Message) {
	switch msg.Type {

	case "call_request":
		// Check if target user is busy first
		if h.activeCall[msg.To] != "" {
			sender.send <- Message{Type: "busy", To: sender.username}
			return
		}

		// Mark both users as in a call
		h.activeCall[sender.username] = msg.CallID
		h.activeCall[msg.To] = msg.CallID

		// Try to deliver via WebSocket first (if online)
		target, online := h.clients[msg.To]
		if online {
			// User is connected - send WebSocket message
			target.send <- Message{
				Type:   "incoming_call",
				From:   sender.username,
				CallID: msg.CallID,
			}
		} else {
			// User is OFFLINE - send FCM push notification
			fcmToken, hasToken := h.GetFCMToken(msg.To)
			if hasToken {
				payload := map[string]string{
					"type":         "call_initiate",
					"uuid":         msg.CallID,
					"callerName":   sender.username,
					"callerHandle": sender.username,
				}
				go sendPushNotification(fcmToken, payload)
			}
		}

	case "call_accept", "call_reject", "sdp_offer", "sdp_answer", "ice_candidate":
		target, ok := h.clients[msg.To]
		if !ok {
			return
		}
		msg.From = sender.username
		target.send <- msg

	case "call_end":
		callID := msg.CallID
		for user, cID := range h.activeCall {
			if cID == callID {
				delete(h.activeCall, user)
			}
		}
		for _, c := range h.clients {
			c.send <- Message{Type: "call_ended", CallID: callID}
		}
	}
}
