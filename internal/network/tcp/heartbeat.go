package tcp

import (
	"encoding/json"
	"time"
)

const (
	heartbeatInterval = 30 * time.Second
	heartbeatTimeout  = 60 * time.Second
)

type HeartbeatMessage struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *GameServer) startHeartbeat(client *Client) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if time.Since(client.LastPing) > heartbeatTimeout {
				s.removeClient(client)
				return
			}

			msg := HeartbeatMessage{
				Type:      "ping",
				Timestamp: time.Now(),
			}

			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}

			select {
			case client.SendChan <- data:
			default:
				// Channel full, skip
			}

		case <-s.done:
			return
		}
	}
}
