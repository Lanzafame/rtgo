//    Title: message.go
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

// Message defines the structure of incoming JSON messages
// that do not perform DB functions.
type Message struct {
	Room    string `json:"room"`
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

// DBMessage defines the structure of incoming JSON messages
// that do perform DB function.
type DBMessage struct {
	DB    string `json:"db"`
	Table string `json:"table"`
	Key   string `json:"key"`
	Data  string `json:"data"`
}
