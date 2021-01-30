package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/samanyu6/goChat/chat"
	uuid "github.com/satori/go.uuid"
)

func main() {
	fmt.Println("Starting application")
	go chat.Manager.Start()
	http.HandleFunc("/talk", talk)
	http.ListenAndServe(":8000", nil)
}

func talk(res http.ResponseWriter, req *http.Request) {
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}).Upgrade(res, req, nil)

	if err != nil {
		http.NotFound(res, req)
		return
	}

	client := &chat.Client{Id: uuid.NewV4().String(), Socket: conn, Send: make(chan []byte)}
	chat.Manager.Register <- client

	go client.Read()
	go client.Write()
}
