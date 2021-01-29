package chat

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/samanyu6/goChat/data"
)

type ClientManager data.ClientManager
type Client data.Client

var Manager = ClientManager{
	Broadcast:  make(chan []byte),
	Register:   make(chan *data.Client),
	Unregister: make(chan *data.Client),
	Clients:    make(map[*data.Client]bool),
}

func (Manager *ClientManager) Start() {
	for {
		select {

		// create a new socket conn
		case Conn := <-Manager.Register:
			Manager.Clients[Conn] = true
			jsonMsg, _ := json.Marshal(&data.Message{Content: "New Socket Connection Created. \n"})
			fmt.Println(jsonMsg)
			// Manager.send(jsonMsg)

			// remove socket conn for specific client
		case Conn := <-Manager.Unregister:
			if _, ok := Manager.Clients[Conn]; ok {
				close(Conn.Send)
				delete(Manager.Clients, Conn)
				jsonMsg, _ := json.Marshal(&data.Message{Content: "Socket Disconnected."})
				fmt.Println(jsonMsg)
				// Manager.send(jsonMsg, Conn)
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

func (Manager *ClientManager) Send(message []byte, ignore *data.Client) {
	for Conn := range Manager.Clients {
		if Conn != ignore {
			Conn.Send <- message
		}
	}
}

// workaround since you can;t import data types not defined within this package.
func (c *Client) Read() {
	c1 := data.Client(*c)
	defer func() {
		Manager.Unregister <- &c1
		c1.Socket.Close()
	}()

	for {
		_, message, err := c1.Socket.ReadMessage()
		if err != nil {
			Manager.Unregister <- &c1
			c1.Socket.Close()
			break
		}

		jsonMsg, _ := json.Marshal(&data.Message{Sender: c1.Id, Content: string(message)})
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
