//    Title: wsrooms.js
//    Author: Jon Cody
//    Year: 2014
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


(function (global) {
    'use strict';

/**
 * WSRooms
 * Contsructor of WSRooms
 * @param {String} url
 * @param {String || Array} protocol
 */
    function WSRooms(url) {
        if (global.WebSocket && typeof url === 'string') {
            eventEmitter(this);
            this.open = false;
            this.id = null;
            this.room = 'root';
            this.rooms = {};
            this.socket = new WebSocket(url);
            this.socket.onmessage = this.onmessage.bind(this);
            this.socket.onclose = this.onclose.bind(this);
            this.socket.onerror = this.onerror.bind(this);
        }
    }

/**
 * WSRooms.send
 * Send a message.
 * @param {String} room
 * @param {String} event
 * @param {String || Array || Object || Boolean || Number || null} payload
 */
    WSRooms.prototype.send = function send(room, event, payload) {
        var data = {};

        if (typeof event !== undefined && room && typeof room === 'string' && payload === undefined) {
            payload = event;
            event = room;
            room = 'root';
        }
        if (this.open && typeof room === 'string' && typeof event === 'string' && typeof payload !== undefined && (room === 'root' || this.rooms.hasOwnProperty(room))) {
            data.room = room;
            data.event = event;
            if (typeof payload !== 'string') {
                try {
                    payload = JSON.stringify(payload);
                } catch (ignore) {}
            }
            data.payload = payload;
            this.socket.send(JSON.stringify(data));
        }
    };

/**
 * WSRooms.handleMessage
 * Internal method used to handle the contents of a message after it has been received.
 * @param {String} room
 * @param {String} event
 * @param {String || Array || Object || Boolean || Number || null} payload
 */
    WSRooms.prototype.handleMessage = function handleMessage(room, event, payload) {
        if (event && typeof event === 'string' && room && typeof room === 'string' && typeof payload !== undefined) {
            if (room === 'root') {
                if (event === 'join') {
                    this.id = payload;
                    this.open = true;
                    this.emit('open');
                } else if (event === 'leave') {
                    this.socket.close();
                } else {
                    this.emit(event, payload);
                }
            } else if (this.rooms.hasOwnProperty(room)) {
                if (event === 'join') {
                    this.rooms[room].open = true;
                    this.rooms[room].emit('open');
                } else if (event === 'leave') {
                    this.rooms[room].open = false;
                    this.rooms[room].emit('close');
                    delete this.rooms[room];
                } else {
                    this.rooms[room].emit(event, payload);
                }
            }
        }
    };

/**
 * WSRooms.onmessage
 * Called when a message is received.
 * @param {Event} e
 */
    WSRooms.prototype.onmessage = function onmessage(e) {
        var data = e.data,
            room,
            event,
            payload;

        try {
            data = JSON.parse(data);
        } catch (ignore) {
            return console.log("Invalid message format: Not JSON");
        }
        if (typeof data.payload !== undefined) {
            try {
                payload = JSON.parse(data.payload);
            } catch (ignore) {
                payload = data.payload;    
            }
        } else {
            payload = {};
        }
        event = data.event;
        room = data.room;
        this.handleMessage(room, event, payload);
    };

/**
 * WSRooms.join
 * Join a room. Returns an eventEmitter object with the methods 'send' and 'leave'.
 * @param {String} room
 * @return {Object} sock
 */
    WSRooms.prototype.join = function join(room) {
        var sock = {};

        if (this.open && room && typeof room === 'string' && !this.rooms.hasOwnProperty(room)) {
            eventEmitter(sock);
            sock.id = this.id;
            sock.open = false;
            sock.room = room;
            sock.send = this.send.bind(this, room);
            sock.leave = this.leave.bind(this, room);
            sock.close = sock.leave;
            this.rooms[room] = sock;
            this.send(room, 'join', null);
        }
        return sock;
    };

/**
 * WSRooms.leave / WSRooms.close
 * Leave a room.
 * @param {String} room
 */
    WSRooms.prototype.leave = function leave(room) {
        this.send(room, 'leave', null)
    };
    WSRooms.prototype.close = WSRooms.prototype.leave;

/**
 * WSRooms.close
 * Called when an instance of WSRooms is closed.
 */
    WSRooms.prototype.onclose = function onclose() {
        Object.keys(this.rooms).forEach(function (room) {
            this.rooms[room].emit('close');
            delete this.rooms[room];
        }, this);
        this.open = false;
        this.emit('close');
    };

/**
 * WSRooms.onerror
 * Called when an error occurs on an instance of WSRooms.
 * @param {Event} e
 */
    WSRooms.prototype.onerror = function onerror(e) {
        this.emit('error', e);
    };

    global.wsrooms = function wsrooms(url) {
        return new WSRooms(url);
    };

}(this || window));
