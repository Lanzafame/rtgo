//    Title: conn.go
//    Author: JD
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

package rtgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Conn struct {
	app       *App
	socket    *websocket.Conn
	id        string
	send      chan []byte
	rooms     map[string]*Room
	privilege string
}

// SendView sends the view matching requested path.
func (c *Conn) SendView(path string) {
	var doc bytes.Buffer
	var err error
	route := c.app.FindRoute(path)
	if _, ok := route["template"]; !ok {
		log.Println("No template for the specified path: ", path)
		return
	}
	collection := make([]interface{}, 0)
	if _, ok := route["table"]; ok {
		for _, db := range c.app.DBManager {
			collection, err = db.GetAllObjs(route["table"])
			if err != nil {
				continue
			}
			break
		}
	}
	c.app.Templates.ExecuteTemplate(&doc, route["template"], collection)
	response := map[string]interface{}{
		"room":  "root",
		"event": "response",
		"payload": map[string]string{
			"template":   doc.String(),
			"controller": route["controller"],
		},
	}
	data, err := json.Marshal(&response)
	if err != nil {
		log.Println("error encoding json: ", err)
		return
	}
	c.send <- data
}

// HandleData routes a received message.
// By default, the message is emitted on the WSEmitter.
// It returns an error if any occur.
func (c *Conn) HandleData(data *Message) error {
	switch data.Event {
	default:
		c.app.Emitter.Emit(data.Event, c, data)
	case "join":
		c.Join(data.Room)
	case "leave":
		c.Leave(data.Room)
	case "request":
		c.SendView(data.Payload)
	case "getObj":
		if c.privilege != "admin" {
		    return nil
		}
		payload := &DBMessage{}
		if err := json.Unmarshal([]byte(data.Payload), payload); err != nil {
			return err
		}
		if _, exists := c.app.DBManager[payload.DB]; !exists {
			return errors.New("Database does not exist.")
		}
		obj, err := c.app.DBManager[payload.DB].GetObj(payload.Table, payload.Key)
		if err != nil {
			return err
		}
		newdata := &Message{
			Room:    "root",
			Event:   "gotObj",
			Payload: obj.(string),
		}
		c.Emit(newdata)
	case "insertObj":
		if c.privilege != "admin" {
			return nil
		}
		payload := &DBMessage{}
		err := json.Unmarshal([]byte(data.Payload), payload)
		if err != nil {
			return err
		}
		if _, exists := c.app.DBManager[payload.DB]; !exists {
			return errors.New("Database does not exist.")
		}
		if err := c.app.DBManager[payload.DB].InsertObj(payload.Table, payload.Key, payload.Data); err != nil {
			return err
		}
	case "deleteObj":
		if c.privilege != "admin" {
			return nil
		}
		payload := &DBMessage{}
		err := json.Unmarshal([]byte(data.Payload), payload)
		if err != nil {
			return err
		}
		if _, exists := c.app.DBManager[payload.DB]; !exists {
			return errors.New("Database does not exist.")
		}
		if err := c.app.DBManager[payload.DB].DeleteObj(payload.Table, payload.Key); err != nil {
			return err
		}
	}
	return nil
}

// ReadPump reads and parses incoming messages before passing them to HandleData.
func (c *Conn) ReadPump() {
	defer func() {
		for _, room := range c.rooms {
			room.leave <- c
		}
		c.socket.Close()
	}()
	c.socket.SetReadLimit(maxMessageSize)
	c.socket.SetReadDeadline(time.Now().Add(pongWait))
	c.socket.SetPongHandler(func(string) error {
		c.socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		data := &Message{}
		if err := c.socket.ReadJSON(data); err != nil {
			if err != io.EOF {
				log.Println("error parsing incoming message:", err)
			}
			break
		}
		if err := c.HandleData(data); err != nil {
			log.Println(err)
		}
	}
}

// Write writes a message with the given message type and payload to the WebSocket connection.
func (c *Conn) Write(mt int, payload []byte) error {
	c.socket.SetWriteDeadline(time.Now().Add(writeWait))
	return c.socket.WriteMessage(mt, payload)
}

// WritePump pumps messages from a room to the WebSocket connection.
func (c *Conn) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.socket.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.Write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Write(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// Join will cause the WebSocket connection to join a room with name.
func (c *Conn) Join(name string) {
	var room *Room
	if _, ok := c.app.RoomManager[name]; ok {
		room = c.app.RoomManager[name]
	} else {
		room = c.app.NewRoom(name)
	}
	room.Join(c)
	c.rooms[name] = room
}

// Leave removes the WebSocket connection from a room with name.
func (c *Conn) Leave(name string) {
	if room, ok := c.app.RoomManager[name]; ok {
		room.Leave(c)
		delete(c.rooms, room.name)
	}
}

// Emit sends a message to all connections in a room specified in payload.
func (c *Conn) Emit(payload *Message) {
	if room, ok := c.app.RoomManager[payload.Room]; ok {
		room.Emit(payload)
	}
}
