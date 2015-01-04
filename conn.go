package main

import (
	"code.google.com/p/go-uuid/uuid"
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

type RTConn struct {
	socket    *websocket.Conn
	id        string
	send      chan []byte
	rooms     map[string]*RTRoom
	privilege string
}

var ConnManager = make(map[string]*RTConn)

// handleData handles incoming JSON data received by a WebSocket connection, or an RTConn instance.
func (c *RTConn) handleData(data *Message) error {
	switch data.Event {
	default:
		c.Emit(data)
	case "join":
		c.Join(data.Room)
	case "leave":
		c.Leave(data.Room)
	case "request":
		SendView(c, data.Payload)
	case "getObj":
		//	if c.privilege != "admin" {
		//		return
		//	}
		payload := &DBMessage{}
		if err := json.Unmarshal([]byte(data.Payload), payload); err != nil {
			return err
		}
		if _, exists := DBManager[payload.DB]; !exists {
			return errors.New("Database does not exist.")
		}
		obj, err := DBManager[payload.DB].GetObj(payload.Table, payload.Key)
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
		//	if c.privilege != "admin" {
		//		return
		//	}
		payload := &DBMessage{}
		err := json.Unmarshal([]byte(data.Payload), payload)
		if err != nil {
			return err
		}
		if _, exists := DBManager[payload.DB]; !exists {
			return errors.New("Database does not exist.")
		}
		if err := DBManager[payload.DB].InsertObj(payload.Table, payload.Key, payload.Data); err != nil {
			return err
		}
	case "deleteObj":
		//	if c.privilege != "admin" {
		//		return
		//	}
		payload := &DBMessage{}
		err := json.Unmarshal([]byte(data.Payload), payload)
		if err != nil {
			return err
		}
		if _, exists := DBManager[payload.DB]; !exists {
			return errors.New("Database does not exist.")
		}
		if err := DBManager[payload.DB].DeleteObj(payload.Table, payload.Key); err != nil {
			return err
		}
	}
	return nil
}

// readPump reads and parses incoming JSON blobs before passing them to handleData.
func (c *RTConn) readPump() {
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
		if err := c.handleData(data); err != nil {
			log.Println(err)
		}
	}
}

// write writes data to the client over the websocket connection.
func (c *RTConn) write(mt int, payload []byte) error {
	c.socket.SetWriteDeadline(time.Now().Add(writeWait))
	return c.socket.WriteMessage(mt, payload)
}

// writePump
func (c *RTConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.socket.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// Join will cause a connection to join a specified room.
func (c *RTConn) Join(name string) {
	var room *RTRoom
	if _, ok := RoomManager[name]; ok {
		room = RoomManager[name]
	} else {
		room = NewRoom(name)
	}
	room.Join(c)
	c.rooms[name] = room
}

// Leave removes a connection from a specified room.
func (c *RTConn) Leave(name string) {
	if room, ok := RoomManager[name]; ok {
		room.Leave(c)
		delete(c.rooms, room.name)
	}
}

// Emit sends an instance of Message to all clients in the specified room.
func (c *RTConn) Emit(payload *Message) {
	if room, ok := RoomManager[payload.Room]; ok {
		room.Emit(payload)
	}
}

// NewConnection returns an instance of RTConn upon upgrading the incoming request.
func NewConnection(w http.ResponseWriter, r *http.Request) *RTConn {
	cookie := ReadCookieHandler(w, r, config.Cookiename)
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	c := &RTConn{
		socket:    socket,
		id:        uuid.New(),
		send:      make(chan []byte, 256),
		rooms:     make(map[string]*RTRoom),
		privilege: cookie["privilege"],
	}
	ConnManager[c.id] = c
	return c
}

// SocketHandler handles incoming WebSocket requests by calling NewConnection.
func SocketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	c := NewConnection(w, r)
	if c != nil {
		go c.writePump()
		c.Join("root")
		c.readPump()
	}
}
