//    Title: room.go
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
	"encoding/json"
	"log"
)

type Room struct {
	app     *App
	name    string
	members map[*Conn]bool
	stop    chan bool
	join    chan *Conn
	leave   chan *Conn
	send    chan []byte
}

// Start activates the room.
func (r *Room) Start() {
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
					Event:   "join",
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

// Stop deactivates the room.
func (r *Room) Stop() {
	r.stop <- true
}

// Join will add a connection to the room.
func (r *Room) Join(c *Conn) {
	r.join <- c
}

// Leave will remove a connection from a room.
func (r *Room) Leave(c *Conn) {
	r.leave <- c
}

// Emit will send a message to all connections in the room.
func (r *Room) Emit(payload *Message) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
		return
	}
	r.send <- data
}
