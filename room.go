package rtgo

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

var RoomManager = make(map[string]*RTRoom)

func (r *RTRoom) Start() {
	for {
		select {
		case c := <-r.join:
			payload := &Message{
				Room:    r.name,
				Event:   "join",
				Payload: c.id,
			}
			data, err := json.Marshal(payload)
			if err != nil {
				log.Println(err)
				break
			}
			c.send <- data
			r.members[c] = true
		case c := <-r.leave:
			if _, ok := r.members[c]; ok {
				payload := &Message{
					Room:    r.name,
					Event:   "leave",
					Payload: c.id,
				}
				data, err := json.Marshal(payload)
				if err != nil {
					log.Println(err)
					break
				}
				c.send <- data
				delete(r.members, c)
			}
		case data := <-r.send:
			for c := range r.members {
				select {
				case c.send <- data:
				default:
					close(c.send)
					delete(r.members, c)
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

func (r *RTRoom) Join(c *RTConn) {
	r.join <- c
}

func (r *RTRoom) Leave(c *RTConn) {
	r.leave <- c
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
