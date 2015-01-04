package main

import (
	"encoding/json"
	"log"
)

type RTRoom struct {
	name    string
	members map[*RTConn]bool
	stop    chan bool
	join    chan *RTConn
	leave   chan *RTConn
	send    chan []byte
}

func (r *RTRoom) Start() {
	for {
		select {
		case conn := <-r.join:
			payload := &Message{
				Room:    r.name,
				Event:   "join",
				Payload: conn.id,
			}
			data, err := json.Marshal(payload)
			if err != nil {
				log.Println(err)
				break
			}
			conn.send <- data
			r.members[conn] = true
		case conn := <-r.leave:
			if _, ok := r.members[conn]; ok {
				payload := &Message{
					Room:    r.name,
					Event:   "leave",
					Payload: conn.id,
				}
				data, err := json.Marshal(payload)
				if err != nil {
					log.Println(err)
					break
				}
				conn.send <- data
				delete(r.members, conn)
			}
		case data := <-r.send:
			for conn := range r.members {
				select {
				case conn.send <- data:
				default:
					close(conn.send)
					delete(r.members, conn)
				}
			}
		case <-r.stop:
			return
		}
	}
}

func (r *RTRoom) Stop() {
	r.stop <- true
}

func (r *RTRoom) Join(conn *RTConn) {
	r.join <- conn
}

func (r *RTRoom) Leave(conn *RTConn) {
	r.leave <- conn
}

func (r *RTRoom) Emit(payload *Message) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		return
	}
	r.send <- data
}

func NewRoom(name string) *RTRoom {
	r := &RTRoom{
		name:    name,
		members: make(map[*RTConn]bool),
		stop:    make(chan bool),
		join:    make(chan *RTConn),
		leave:   make(chan *RTConn),
		send:    make(chan []byte, 256),
	}
	RoomManager[name] = r
	go r.Start()
	return r
}
