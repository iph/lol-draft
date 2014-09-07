package net

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"log"
	"sync"
)

type Server struct {
	Connections   []*Connection
	in            chan *Packet
	out           chan *Packet
	Receive       func(*Packet)
	NewConnection func(*Connection)
	id_counter    int
	mutex         sync.Mutex
}

// Allows for unique connection ids.
func (s *Server) AssignUniqueId() int {
	id := 0
	s.mutex.Lock()
	id = s.id_counter
	s.id_counter = s.id_counter + 1
	s.mutex.Unlock()
	return id

}

func NewServer() *Server {
	return &Server{
		Connections: make([]*Connection, 0),
		in:          make(chan *Packet),
		out:         make(chan *Packet),
	}
}

// Send a message to all other connections.
func (s *Server) Broadcast(message Message, from_conn *Connection) {
	for _, conn := range s.Connections {

		packet := Packet{
			To:      conn,
			From:    from_conn,
			Message: message,
		}
		s.out <- &packet
	}

}

// starts the daemons up to listen for server messages.
func (s *Server) Run() {
	go s.ListenDaemon()
	go s.SendDaemon()
}

func (s *Server) Send(message Message, to *Connection) {
	packet := Packet{
		To:      to,
		Message: message,
	}
	s.out <- &packet
}

// Sits in its own world and listens. Waits in the deep below...always waiting..always listening.
func (s *Server) ListenDaemon() {
	for {
		select {
		case packet := <-s.in:
			s.Receive(packet)
			log.Printf("Packet received: %v\n", packet)
		}
	}
}

// Any message queued to be sent will go through here and be sent properly.
func (s *Server) SendDaemon() {
	for {
		select {
		case packet := <-s.out:
			if !packet.To.Dead {
				data, _ := json.Marshal(packet.Message)
				err := websocket.Message.Send(packet.To.Socket, data)
				if err != nil {
					packet.To.Dead = true
					log.Printf("Failure sending packet: %v", err)
				}
			}
		}
	}
}

// Adds a connection to the server.
func (s *Server) AddConnection(websock *websocket.Conn) *Connection {
	connection := Connection{websock, s.AssignUniqueId(), false}
	s.Connections = append(s.Connections, &connection)
	log.Printf("Connection added")
	s.NewConnection(&connection)
	return &connection
}
