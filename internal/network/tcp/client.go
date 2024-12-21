package tcp

import (
	"net"
	"time"

	"github.com/TaViKhang/pokecat-n-pokebat/internal/models"
)

type Client struct {
	Conn     net.Conn
	Player   *models.Player
	SendChan chan []byte
	LastPing time.Time
	Session  *SessionState
}

func NewClient(conn net.Conn, player *models.Player) *Client {
	return &Client{
		Conn:     conn,
		Player:   player,
		SendChan: make(chan []byte, 100),
		Session: &SessionState{
			LastActivity: time.Now(),
		},
	}
}
