//    Title: eventEmitter.js
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
 * eventEmitter
 * Turn an object into an event emitter.
 * @params {Object} emitter
 * @return {Object} emitter
 */
    function eventEmitter(emitter) {
        emitter = emitter && typeof emitter === 'object' ? emitter : {};
        emitter.events = {};
        emitter.addListener = function (type, listener) {
            var list = emitter.events[type];

            if (typeof listener === 'function') {
                if (emitter.events.newListener) {
                    emitter.emit('newListener', type, typeof listener.listener === 'function' ?
                            listener.listener : listener);
                }
                if (!list) {
                    emitter.events[type] = [listener];
                } else {
                    emitter.events[type].push(listener);
                }
            }
            return emitter;
        };

/**
 * emmitter.on
 * Add an event listener.
 * @param {String} type
 * @param {Function} listener
 * @return {Object} emitter
 */
        emitter.on = emitter.addListener;

/**
 * emmitter.once
 * Add an event listener.
 * @param {String} type
 * @param {Function} listener
 * @return {Object} emitter
 */
        emitter.once = function (type, listener) {
            function g() {
                emitter.removeListener(type, g);
                listener.apply(emitter, arguments);
            }
            if (typeof listener === 'function') {
                g.listener = listener;
                emitter.on(type, g);
            }
            return emitter;
        };

/**
 * emmitter.removeListener
 * Remove an event listener.
 * @param {String} type
 * @param {Function} listener
 * @return {Object} emitter
 */
        emitter.removeListener = function (type, listener) {
            var list = emitter.events[type],
                position = -1,
                i;

            if (typeof listener === 'function' && list) {
                for (i = list.length - 1; i >= 0; i -= 1) {
                    if (list[i] === listener || (list[i].listener && list[i].listener === listener)) {
                        position = i;
                        break;
                    }
                }
                if (position >= 0) {
                    if (list.length === 1) {
                        delete emitter.events[type];
                    } else {
                        list.splice(position, 1);
                    }
                    if (emitter.events.removeListener) {
                        emitter.emit('removeListener', type, listener);
                    }
                }
            }
            return emitter;
        };

/**
 * emmitter.off
 * Remove an event listener.
 * @param {String} type
 * @param {Function} listener
 */
        emitter.off = emitter.removeListener;

/**
 * emmitter.removeAllListeners
 * Remove all event listeners.
 * @param {String} type
 * @return {Object} emitter
 */
        emitter.removeAllListeners = function (type) {
            var list,
                i;

            if (!emitter.events.removeListener) {
                if (!type) {
                    emitter.events = {};
                } else {
                    delete emitter.events[type];
                }
            } else if (!type) {
                Object.keys(emitter.events).forEach(function (key) {
                    if (key !== 'removeListener') {
                        emitter.removeAllListeners(key);
                    }
                });
                emitter.removeAllListeners('removeListener');
                emitter.events = {};
            } else {
                list = emitter.events[type];
                for (i = list.length - 1; i >= 0; i -= 1) {
                    emitter.removeListener(type, list[i]);
                }
                delete emitter.events[type];
            }
            return emitter;
        };

/**
 * emitter.listeners
 * List all listeners of a specified type.
 * @param {String} type
 * @return {Array} list
 */
        emitter.listeners = function (type) {
            var list = [];

            if (type) {
                if (emitter.events[type]) {
                    list = emitter.events[type];
                }
            } else {
                Object.keys(emitter.events).forEach(function (key) {
                    list.push(emitter.events[key]);
                });
            }
            return list;
        };

/**
 * emitter.emit
 * Emit an event.
 * @param {String} type
 * @return {Object} emitter
 */
        emitter.emit = function (type) {
            var list = emitter.events[type],
                bool = false,
                args = [],
                length,
                i;

            if (list) {
                length = arguments.length;
                for (i = 1; i < length; i += 1) {
                    args[i - 1] = arguments[i];
                }
                length = list.length;
                for (i = 0; i < length; i += 1) {
                    list[i].apply(emitter, args);
                }
                bool =  true;
            }
            return bool;
        };

        return emitter;
    }

    global.eventEmitter = eventEmitter;

}(this));
