package chat

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

var Manager = ClientManager{
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[*Client]bool),
}

func (Manager *ClientManager) Start() {
	for {
		select {

		// create a new socket conn
		case Conn := <-Manager.Register:
			Manager.Clients[Conn] = true
			jsonMsg, _ := json.Marshal(&Message{Content: "New Socket Connection Created. \n"})
			fmt.Println(jsonMsg)
			Manager.Send(jsonMsg, Conn)

			// remove socket conn for specific client
		case Conn := <-Manager.Unregister:
			if _, ok := Manager.Clients[Conn]; ok {
				close(Conn.Send)
				delete(Manager.Clients, Conn)
				jsonMsg, _ := json.Marshal(&Message{Content: "Socket Disconnected."})
				fmt.Println(jsonMsg)
				Manager.Send(jsonMsg, Conn)
			}

		case msg := <-Manager.Broadcast:
			for Conn := range Manager.Clients {
				select {
				case Conn.Send <- msg:
				default:
					{
						close(Conn.Send)
						delete(Manager.Clients, Conn)
					}
				}
			}
		}
	}
}

func (Manager *ClientManager) Send(message []byte, ignore *Client) {
	for Conn := range Manager.Clients {
		if Conn != ignore {
			Conn.Send <- message
		}
	}
}

// workaround since you can;t import data types not defined within this package.
func (c *Client) Read() {
	defer func() {
		Manager.Unregister <- c
		c.Socket.Close()
	}()

	for {
		_, message, err := c.Socket.ReadMessage()
		if err != nil {
			Manager.Unregister <- c
			c.Socket.Close()
			break
		}

		jsonMsg, _ := json.Marshal(&Message{Sender: c.Id, Content: string(message)})
		Manager.Broadcast <- jsonMsg
	}
}

func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}
