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

func (conn *RTConn) handleData(data *Message) error {
	switch data.Event {
	case "join":
		conn.Join(data.Room)
	case "leave":
		conn.Leave(data.Room)
	case "request":
		SendView(conn, data.Payload)
	case "getObj":
		//	if conn.privilege != "admin" {
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
		obj, err := DBManager[payload.DB].GetObj(payload.Table, payload.Key)
		if err != nil {
			return err
		}
		newdata := &Message{
			Room:    "root",
			Event:   "gotObj",
			Payload: obj.(string),
		}
		conn.Emit(newdata)
	case "insertObj":
		//	if conn.privilege != "admin" {
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
		//	if conn.privilege != "admin" {
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
	default:
		conn.Emit(data)
	}
	return nil
}

func (conn *RTConn) readPump() {
	defer func() {
		for _, room := range conn.rooms {
			room.leave <- conn
		}
		conn.socket.Close()
	}()
	conn.socket.SetReadLimit(maxMessageSize)
	conn.socket.SetReadDeadline(time.Now().Add(pongWait))
	conn.socket.SetPongHandler(func(string) error {
		conn.socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		data := &Message{}
		err := conn.socket.ReadJSON(data)
		if err != nil {
			if err != io.EOF {
				log.Println("error parsing incoming message:", err)
			}
			break
		}
		if err := conn.handleData(data); err != nil {
			log.Println(err)
		}
	}
}

func (conn *RTConn) write(mt int, payload []byte) error {
	conn.socket.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.socket.WriteMessage(mt, payload)
}

func (conn *RTConn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.socket.Close()
	}()
	for {
		select {
		case msg, ok := <-conn.send:
			if !ok {
				conn.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.write(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			if err := conn.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (conn *RTConn) Join(name string) {
	var room *RTRoom
	if _, ok := RoomManager[name]; ok {
		room = RoomManager[name]
	} else {
		room = NewRoom(name)
	}
	room.Join(conn)
	conn.rooms[name] = room
}

func (conn *RTConn) Leave(name string) {
	if room, ok := RoomManager[name]; ok {
		room.Leave(conn)
		delete(conn.rooms, room.name)
	}
}

func (conn *RTConn) Emit(payload *Message) {
	if room, ok := RoomManager[payload.Room]; ok {
		room.Emit(payload)
	}
}

func NewConnection(w http.ResponseWriter, r *http.Request) *RTConn {
	cookie := ReadCookieHandler(w, r, "rtgo")
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	conn := &RTConn{
		socket:    socket,
		id:        uuid.New(),
		send:      make(chan []byte, 256),
		rooms:     make(map[string]*RTRoom),
		privilege: cookie["privilege"],
	}
	ConnManager[conn.id] = conn
	return conn
}

func SocketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn := NewConnection(w, r)
	if conn != nil {
		go conn.writePump()
		conn.Join("root")
		conn.readPump()
	}
}
