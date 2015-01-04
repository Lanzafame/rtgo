//    Title: rtgo.js
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

    var wsurl = (global.location.protocol === 'http:' ? 'ws://' : 'wss://') + global.location.host + '/ws',
        dbs = ['riak', 'postgresql', 'mysql', 'sqlite3'];

/**
 * RTGo
 * Contructor of RTGo.
 * @params {String} url
 */
    function RTGo(url) {
        if (typeof url === 'string') {
            this.controllers = {};
            this.hash = '';
            this.view = document.querySelector('[data-rt-view]');
            this.hrefs = document.querySelectorAll('[data-rt-href]');
            this.socket = wsrooms(url);
            this.socket.on('open', this.onopen.bind(this));
            this.socket.on('close', this.onclose.bind(this));
            this.socket.on('error', this.onerror.bind(this));
            this.socket.on('response', this.onresponse.bind(this));
            global.addEventListener('hashchange', this.onhashchange.bind(this), false);
        }
    }

/**
 * RTGo.requestView
 * Requests a view over the WebSocket connection.
 * @param {String} hash
 */
    RTGo.prototype.requestView = function requestView(hash) {
        this.socket.send('request', hash || this.hash);
    };

/**
 * RTGo.assignHrefs
 * Attach event listeners to all elements with a data-rt-href attribute.
 */
    RTGo.prototype.assignHrefs = function assignHrefs() {
        var hrefs = this.hrefs,
            node,
            x;

        function setup(path) {
            return function (e) {
                global.location.hash = curhash.replace(/(\#)/g, '').replace(/(\/\/)/g, '/');
            };
        }

        if (hrefs && hrefs.length) {
            for (x = 0; x < hrefs.length; x += 1) {
                node = hrefs[x];
                node.addEventListener('click', setup(node.getAttribute('data-rt-href')), false);
            }
        }
    };

/**
 * RTGo.onopen
 * Called when the WebSocket connection is opened. 
 * Requests the view associated with the root path.
 */
    RTGo.prototype.onopen = function onopen() {
        var curhash = global.location.hash;

        if (!curhash) {
            global.location.hash = '/';
        } else {
            global.location.hash = curhash.replace(/(\#)/g, '').replace(/(\/\/)/g, '/');
            this.onhashchange();
        }
    };

/**
 * RTGo.onclose
 * Called when the WebSocket connection is closed.
 */
    RTGo.prototype.onclose = function onclose() {
        console.log('closed socket');
    };

/**
 * RTGo.onerror
 * Called when the WebSocket connection encounters an error.
 */
    RTGo.prototype.onerror = function onerror(e) {
        console.log('socket error:', e);
    };

/**
 * RTGo.onresponse
 * Called when a response is received from RTGo.requestView;
 * data.template is placed in the tag with the data-rt-view="" attribute;
 * data.controller is the name of the controller function which will be executed.
 * @param {Object} data
 */
    RTGo.prototype.onresponse = function onresponse(data) {
        var template = data.template,
            controller = data.controller;

        if (this.view) {
            this.view.innerHTML = template;
        }
        if (this.controllers.hasOwnProperty(controller) && typeof this.controllers[controller] === 'function') {
            this.controllers[controller]();
        }
        this.hrefs = document.querySelectorAll('[data-rt-href]');
        this.assignHrefs();
    };

/**
 * RTGo.onhashchange
 * Called when the URL hash changes; requesting a new view.
 * @param {Event Object} e
 */
    RTGo.prototype.onhashchange = function onhashchange(e) {
        var curhash = global.location.hash;

        if (curhash && this.hash !== curhash) {
            this.hash = curhash.replace(/\#/g, '').replace(/\/\//g, '/');
            this.requestView();
        }
    };

/**
 * checkParams
 * Checks the parameters give to the below functions
 * which query the databases on the server.
 */
    function checkParams(db, table, key) {
        return db && typeof db === 'string' && dbs.indexOf(db) !== -1 &&
                table && typeof table === 'string' &&
                key && typeof key === 'string';
    }

/**
 * RTGo.getObj
 * @param {String} db
 * @param {String} table
 * @param {String} key
 */
    RTGo.prototype.getObj = function getObj(db, table, key) {
        if (checkParams(db, table, key)) {
            this.socket.send("getObj", {
                db: db,
                table: table,
                key: key
            });
        }
    };

/**
 * RTGo.insertObj
 * @param {String} db
 * @param {String} table
 * @param {String} key
 * @param {String || Number || Boolean || Array || Object || null} data
 */
    RTGo.prototype.insertObj = function insertObj(db, table, key, data) {
        if (checkParams(db, table, key)) {
            try {
                data = JSON.stringify(data);
            } catch (ignore) {}
            this.socket.send("insertObj", {
                db: db,
                table: table,
                key: key,
                data: data
            });
        }
    };

/**
 * RTGo.deleteObj
 * @param {String} db
 * @param {String} table
 * @param {String} key
 */
    RTGo.prototype.deleteObj = function deleteObj(db, table, key) {
        if (checkParams(db, table, key)) {
            this.socket.send("deleteObj", {
                db: db,
                table: table,
                key: key
            });
        }
    };

    global.rtgo = new RTGo(wsurl);

}(this || window));
