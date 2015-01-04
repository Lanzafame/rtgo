(function (global) {
    'use strict';

    var wsurl = (global.location.protocol === 'http:' ? 'ws://' : 'wss://') + global.location.host + '/ws',
        dbs = ['riak', 'postgresql', 'mysql', 'sqlite3'];

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

    RTGo.prototype.requestView = function requestView() {
        this.socket.send('request', this.hash);
    };

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

    RTGo.prototype.onopen = function onopen() {
        var curhash = global.location.hash;

        if (!curhash) {
            global.location.hash = '/';
        } else {
            global.location.hash = curhash.replace(/(\#)/g, '').replace(/(\/\/)/g, '/');
            this.onhashchange();
        }
    };

    RTGo.prototype.onclose = function onclose() {
        console.log('closed socket');
    };

    RTGo.prototype.onerror = function onerror(e) {
        console.log('socket error:', e);
    };

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

    RTGo.prototype.onhashchange = function onhashchange(e) {
        var curhash = global.location.hash;

        if (curhash && this.hash !== curhash) {
            this.hash = curhash.replace(/\#/g, '').replace(/\/\//g, '/');
            this.requestView();
        }
    };

    function checkParams(db, table, key) {
        return db && typeof db === 'string' && dbs.indexOf(db) !== -1 &&
                table && typeof table === 'string' &&
                key && typeof key === 'string';
    }

    RTGo.prototype.getObj = function getObj(db, table, key) {
        if (checkParams(db, table, key)) {
            this.socket.send("getObj", {
                db: db,
                table: table,
                key: key
            });
        }
    };

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
