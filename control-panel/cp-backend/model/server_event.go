package model

import "time"

type ServerEvent struct {
	EventType string `json:"eventType"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}


