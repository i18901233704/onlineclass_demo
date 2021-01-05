// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"log"
	"net/http"
	"time"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 1 * time.Second

	// Time allowed to read the next message from the peer.
	pingWait = 600 * time.Second

)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is a middleman between the websocket connection and the proxy.
type Client struct {

	Roomid string

	Userid string
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

func (c *Client) LeaveRoom(){


	msg := map[string]string{}
	msg["cmd"] = "leave"
	msg["roomid"] = c.Roomid
	msg["userid"] = c.Userid
	msg["cli_timestamp"] = fmt.Sprintf("%v_%v",time.Now().UnixNano() / 1e6,1)
	data,_ := json.Marshal(msg)
	select {
		case roomMsgChan <- data :
			fmt.Println("ok leave and broadcast", string(data))
		default:
			fmt.Println("error leave not broadcast", string(data))

	}
}

func (c *Client) SendToRoomMsgChan(msg []byte){

	fmt.Println("recv message from client and send to  roomMsgChan",string(msg))
	select {
		case roomMsgChan <- msg:
			fmt.Println("send to roomMsgChan ok",string(msg))
		default:
			fmt.Println("send to roomMsgChan error",string(msg))
	}
}

func (c *Client) SelfKickoff(){
}

func (c *Client) Destroy(){
}

func (c *Client) ReadMessage() {
	defer func() {
		c.conn.Close()
		close(c.send)
		//c.LeaveRoom()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pingWait))
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			fmt.Println("Client close the  connection")
			return
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		fmt.Println("recv from client",string(message))
		msg := map[string]string{}
		json.Unmarshal(message,&msg)
		roomid := msg["roomid"]
		userid := msg["userid"]
		if msg["cmd"] == "enter" {
			fmt.Println("1 add")
			roomManager.CreateRoom(roomid)
			roomManager.AddUerToRoom(roomid,userid,c)
			fmt.Println("2 add")
		}

		c.SendToRoomMsgChan(message)
	}
}


func (c *Client) SendToClient(msg []byte) {

	//fmt.Println("recv message from room and send to client",string(msg))
	select {
		case c.send <- msg:
			fmt.Println("send to c.send ok",string(msg))
		default:
			fmt.Println("send to c.send error",string(msg))
	}

}

func (c *Client) WriteMessage() {

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				fmt.Println("client closed  and closed chan")
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err ==nil{
				fmt.Println("send message to client  ok",string(message))
			} else{
				fmt.Println("send message to client  error",string(message))
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{conn: conn, send: make(chan []byte, 256)}

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	fmt.Println("create new client")
	go client.WriteMessage()
	go client.ReadMessage()
}
