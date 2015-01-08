package rtgo

import (
	"bytes"
	"code.google.com/p/go-uuid/uuid"
	"encoding/json"
	"errors"
	"github.com/chuckpreslar/emission"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
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

// ConnManager manages, or holds, all existing connections.
var ConnManager = make(map[string]*RTConn)
var WSEmitter = emission.NewEmitter()

// HandleData routes a received message.
// By default, the message is emitted on the WSEmitter.
// It returns an error if any occur.
func (c *RTConn) HandleData(data *Message) error {
	switch data.Event {
	default:
		WSEmitter.Emit(data.Event, c, data)
	case "join":
		c.Join(data.Room)
	case "leave":
		c.Leave(data.Room)
	case "request":
		c.SendView(data.Payload)
	case "getObj":
		// if c.privilege != "admin" {
		//     return nil
		// }
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
		if c.privilege != "admin" {
			return nil
		}
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
		if c.privilege != "admin" {
			return nil
		}
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

// ReadPump reads and parses incoming messages before passing them to HandleData.
func (c *RTConn) ReadPump() {
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
func (c *RTConn) Write(mt int, payload []byte) error {
	c.socket.SetWriteDeadline(time.Now().Add(writeWait))
	return c.socket.WriteMessage(mt, payload)
}

// WritePump pumps messages from a room to the WebSocket connection.
func (c *RTConn) WritePump() {
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

// findRoute loops through all routes attempting to match path.
// It returns the matched route.
func findRoute(path string) map[string]string {
	route := make(map[string]string)
	if _, ok := config.Routes[path]; ok {
		route = config.Routes[path]
	} else {
		for key, _ := range config.Routes {
			if !strings.HasPrefix(key, "^") {
				continue
			}
			reg, err := regexp.Compile(key)
			if err != nil {
				continue
			}
			match := reg.FindStringSubmatch(path)
			if match == nil || len(match) == 0 {
				continue
			}
			for k, val := range config.Routes[key] {
				if !strings.HasPrefix(val, "$") {
					route[k] = val
					continue
				}
				index, err := strconv.Atoi(string(val[1]))
				if err != nil {
					continue
				}
				route[k] = match[index]
			}
		}
	}
	return route
}

// SendView sends the view matching requested path.
func (c *RTConn) SendView(path string) {
	var doc bytes.Buffer
	var err error
	route := findRoute(path)
	if _, ok := route["template"]; !ok {
		log.Println("No template for the specified path: ", path)
		return
	}
	collection := make([]interface{}, 0)
	if _, ok := route["table"]; ok {
		for _, db := range DBManager {
			collection, err = db.GetAllObjs(route["table"])
			if err != nil {
				continue
			}
			break
		}
	}
	config.Templates.ExecuteTemplate(&doc, route["template"], collection)
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

// Join will cause the WebSocket connection to join a room with name.
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

// Leave removes the WebSocket connection from a room with name.
func (c *RTConn) Leave(name string) {
	if room, ok := RoomManager[name]; ok {
		room.Leave(c)
		delete(c.rooms, room.name)
	}
}

// Emit sends a message to all connections in a room specified in payload.
func (c *RTConn) Emit(payload *Message) {
	if room, ok := RoomManager[payload.Room]; ok {
		room.Emit(payload)
	}
}

// NewConnection upgrades an icoming HTTP request, creates a new WebSocket
// connection, and adds it to ConnManager.
// It returns the new connection.
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
