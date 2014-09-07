package net

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/iph/lol-draft/ols"
	"log"
	"strings"
)

type Connection struct {
	Socket *websocket.Conn
	Id     int
	Dead   bool
}

func (c *Connection) Disconnect() {
	c.Socket.Close()
	c.Dead = true
}

// Sits and waits for packages to arrive. Then eats them (sends them to the server to be sorted).
func (c *Connection) Spin(server *Server) {
	for {
		if c.Dead {
			break
		}
		var data string
		err := websocket.Message.Receive(c.Socket, &data)
		log.Printf("Data received %s\n", data)
		if err != nil {
			c.Dead = true
			return
		}

		json_decoder := json.NewDecoder(strings.NewReader(data))
		var message Message
		json_decoder.Decode(&message)
		server.in <- &Packet{To: c, Message: message}
	}

}

type Message struct {
	Token   string
	Type    string
	Payload interface{}
}

type PlayerPayload struct {
	Player ols.Player
}

type BidderInfoPayload struct {
	Points int
	Team   string
}

type BidUpdatePayload struct {
	Bid  string
	Team string
}

type BidWinnerPayload struct {
}

type Packet struct {
	Message Message
	To      *Connection
	From    *Connection
}
