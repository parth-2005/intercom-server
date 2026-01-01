package main

type Message struct {
	Type   string      `json:"type"`
	To     string      `json:"to,omitempty"`
	From   string      `json:"from,omitempty"`
	CallID string      `json:"call_id,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}
