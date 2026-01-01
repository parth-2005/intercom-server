package main

type Message struct {
	Type   string      `json:"type"`
	To     string      `json:"to,omitempty"`
	From   string      `json:"from,omitempty"`
	CallID string      `json:"call_id,omitempty"`
	User   string      `json:"user,omitempty"`
	Online *bool       `json:"online,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}
