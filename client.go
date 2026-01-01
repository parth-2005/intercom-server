package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	username string
	conn     *websocket.Conn
	send     chan Message
}

func (c *Client) readLoop(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			log.Println("read error:", err)
			break
		}
		hub.handleMessage(c, msg)
	}
}

func (c *Client) writeLoop() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			log.Println("write error:", err)
			return
		}
	}
}
