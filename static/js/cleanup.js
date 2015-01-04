//    Title: cleanup.js
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

/*jslint browser: true */
/*globals FileReader */

(function (global) {
    'use strict';

    var module = module;

/**
 * noop
 * Empty function block.
 */
    function noop() {
        return;
    }

/**
 * typeOf
 * Fixing the typeof operator.
 * @param {Any} variable
 * @return {String} type
 */
    function typeOf(variable) {
        var type = typeof variable;

        if (type === 'object') {
            if (Array.isArray(variable)) {
                type = 'array';
            } else if (!variable) {
                type = 'null';
            }
        }
        return type;
    }

/**
 * isBoolean
 * Check if a variable is a boolean.
 * @param {Any} bool
 * @return {Boolean}
 */
    function isBoolean(bool) {
        return typeOf(bool) === 'boolean';
    }

/**
 * isNumber
 * Check if a variable is a number.
 * @param {Any} num
 * @return {Boolean}
 */
    function isNumber(num) {
        return typeOf(num) === 'number';
    }

/**
 * isString
 * Check if a variable is a string.
 * @param {Any} str
 * @return {Boolean}
 */
    function isString(str) {
        return typeOf(str) === 'string';
    }

/**
 * isArray
 * Check if a variable is an array.
 * @param {Any} arr
 * @return {Boolean}
 */
    function isArray(arr) {
        return typeOf(arr) === 'array';
    }

/**
 * isObject
 * Check if a variable is an object.
 * @param {Any} obj
 * @return {Boolean}
 */
    function isObject(obj) {
        return typeOf(obj) === 'object';
    }

/**
 * isFunction
 * Check if a variable is a function.
 * @param {Any} func
 * @return {Boolean}
 */
    function isFunction(func) {
        return typeOf(func) === 'function';
    }

/**
 * isNull
 * Check if a variable is null.
 * @param {Any} nul
 * @return {Boolean}
 */
    function isNull(nul) {
        return typeOf(nul) === 'null';
    }

/**
 * isUndefined
 * Check if a variable is undefined.
 * @param {Any} undef
 * @return {Boolean}
 */
    function isUndefined(undef) {
        return typeOf(undef) === 'undefined';
    }

/**
 * isNode
 * Check if a variable is a Node.
 * @param {Any} node
 * @return {Boolean}
 */
    function isNode(node) {
        return isObject(node) && isString(node.nodeName) && isNumber(node.nodeType);
    }

/**
 * isElement
 * Check if a variable is an HTML Element.
 * @param {Any} node
 * @return {Boolean}
 */
    function isElement(node) {
        return isNode(node) && node.nodeType === 1;
    }

/**
 * isEmptyObject
 * Check if a variable is an empty object.
 * @param {Any} obj
 * @return {Boolean}
 */
    function isEmptyObject(obj) {
        return isObject(obj) && Object.keys(obj).length === 0;
    }

/**
 * isArrayLike
 * Check if a variable is an array-like object.
 * @param {Any} obj
 * @return {Boolean}
 */
    function isArrayLike(obj) {
        var result = false;

        if (isObject(obj) && isNumber(obj.length)) {
            result = Object.keys(obj).every(function (key) {
                return key === 'length' || !global.isNaN(parseInt(key, 10));
            });
        }
        return result;
    }

/**
 * arrayContains
 * Check if an array contains a specified value.
 * @param {Array} arr
 * @param {Any} val
 * @return {Boolean} contains
 */
    function arrayContains(arr, val) {
        var contains = false;

        if (isArray(arr) && arr.indexOf(val) !== -1) {
            contains = true;
        }
        return contains;
    }

/**
 * stringifiable
 * Check if an object can be JSON stringified.
 * @param {Object} obj
 * @return {Boolean}
 */
    function stringifiable(obj) {
        var answer = false;

        try {
            JSON.stringify(obj);
            answer = true;
        } catch (ignore) {}

        return answer;
    }

/**
 * stringify
 * If an object can be JSON stringified, return the object as a JSON string.
 * If the object cannot be JSON stringified, return the object.
 * @param {Object} obj
 * @return {String || obj}
 */
    function stringify(obj) {
        var result = obj;

        try {
            result = JSON.stringify(obj);
            if (result === "{}") {
                result = obj;
            }
        } catch (e) {
            result = obj;
        }
        return result;
    }

/**
 * parsable
 * Check if the passed parameter is JSON parsable.
 * @param {String} str
 * @return {Boolean} parsable
 */
    function parsable(str) {
        var answer = false;

        try {
            JSON.parse(str);
            answer = true;
        } catch (ignore) {}

        return answer;
    }

/**
 * parse
 * If the passed parameter is JSON parsable, parse it.
 * @param {String} str
 * @return {Object || str} result
 */
    function parse(str) {
        var result = str;

        try {
            result = JSON.parse(str);
        } catch (e) {
            result = str;
        }
        return result;
    }

/**
 * toCamelCase
 * Convert a string with hyphens to camelCase.
 * @param {String} str
 * @return {String}
 * this-is-an-example -> thisIsAnExample
 */
    function toCamelCase(str) {
        return isString(str) && str.replace(/-([a-z])/g, function (a) {
            return a[1].toUpperCase();
        });
    }

/**
 * undoCamelCase
 * Convert a camelCase string to a string with hyphens.
 * @param {String} str
 * @return {String}
 * thisIsAnExample -> this-is-an-example
 */
    function undoCamelCase(str) {
        return isString(str) && str.replace(/([A-Z])/g, function (a) {
            return '-' + a.toLowerCase();
        });
    }

/**
 * getbyid
 * A short-hand function for document.getElementById.
 * @param {String} id
 * @return {Node}
 */
    function getbyid(id) {
        return isString(id) && document && document.getElementById(id);
    }

/**
 * select
 * A short-hand function for document.querySelector.
 * @param {String} selector
 * @return {Node}
 */
    function select(selector) {
        return isString(selector) && document && document.querySelector(selector);
    }

/**
 * selectAll
 * A short-hand function for document.querySelectorAll.
 * @param {String} selector
 * @return {Node}
 */
    function selectAll(selector) {
        return isString(selector) && document && document.querySelectorAll(selector);
    }

/**
 * uuid
 * Generate a universally unique identifier.
 * @return {String}
 */
    function uuid() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
            var r = Math.random() * 16 | 0,
                v = c === 'x' ? r : r & 0x3 | 0x8;

            return v.toString(16);
        });
    }

/**
 * supplant
 * Substitute variables within a string noted by brackets: {variable}
 * If the passed object contains a key matching the variable denoted by brackets,
 * its value will replace the variable and brackets.
 * @param {String} s
 * @param {Object} o
 * supplant("Here is an {ex}.", { ex: "example" }) -> "Here is an example."
 */
    function supplant(s, o) {
        return isString(s) && isObject(o) ?
                s.replace(/\{([^{}]*)\}/g, function (a, b) {
                    var r = o[b];

                    return isString(r) || isNumber(r) ? r : a;
                }) : s;
    }

/**
 * inherits
 * Have one object inherit from another object's prototype.
 * @param {Object} ctor
 * @param {Object} superCtor
 * @return {Object} ctor
 */
    function inherits(ctor, superCtor) {
        if (isObject(ctor) && isObject(superCtor)) {
            ctor.superior = superCtor;
            ctor.prototype = Object.create(superCtor.prototype, {
                constructor: {
                    value: ctor,
                    enumberable: false,
                    writable: true,
                    configurable: true
                }
            });
        }
        return ctor;
    }

/**
 * extend
 * Extend one object with another.
 * @param {Object} obj
 * @param {Object} add
 * @param {Boolean} overwrite
 * @return {Object} obj
 */
    function extend(obj, add, overwrite) {
        overwrite = isBoolean(overwrite) ? overwrite : true;
        if (isObject(obj) && isObject(add)) {
            Object.keys(add).forEach(function (key) {
                if (overwrite || !obj.hasOwnProperty(key)) {
                    obj[key] = add[key];
                }
            });
        }
        return obj;
    }

/**
 * each
 * Loop through a variable and pass each of its values as a parameter to a given function.
 * @param {Node, NodeList, Array, Object, Array-Like Object} nodes
 * @param {Function} func
 * @return {Node, NodeList, Array, Object, Array-Like Object} nodes
 */
    function each(nodes, func) {
        var node,
            index,
            length;

        if (nodes && isFunction(func)) {
            if (isNode(nodes)) {
                func.call(nodes, nodes, 0, nodes);
            } else if (isArrayLike(nodes)) {
                length = nodes.length;
                for (index = 0; index < length; index += 1) {
                    node = nodes[index];
                    func.call(nodes, node, index, nodes);
                }
            } else if (isObject(nodes)) {
                Object.keys(nodes).forEach(function (key) {
                    node = nodes[key];
                    func.call(nodes, node, key, nodes);
                });
            } else if (isArray(nodes)) {
                nodes.forEach(func, nodes);
            }
        }
        return nodes;
    }

/**
 * keyboardHandler
 * Assign onkeydown event handlers which are triggered according to a pressed keys' keyCode.
 * @param {Object} options
 * options = {
 *     13: function onEnter() {},
 *     37: function onLeftArrow() {},
 *     38: function onUpArrow() {},
 *     39: function onRightArrow() {},
 *     40: function onDownArrow() {}
 * };
 */
    function keyboardHandler(options) {
        var handlers = {};

        function keyDown(e) {
            var keycode = e.keyCode;

            if (isNumber(keycode) && handlers.hasOwnProperty(keycode)) {
                handlers[keycode](e);
            }
        }

        if (isObject(options)) {
            Object.keys(options).forEach(function (key) {
                var func = options[key];

                key = parseInt(key, 10);
                if (isFunction(func) && isNumber(key) && !global.isNaN(key)) {
                    handlers[key] = func;
                }
            });
            document.addEventListener('onkeydown', keyDown, false);
        }
    }

/**
 * mouseHandler
 * Assign onmousepress event handlers which are triggered according to a pressed keys' keyCode.
 * @param {Object} options
 * options = {
 *     0: function onLeftClick() {},
 *     2: function onRightClick() {}
 * };
 */
    function mouseHandler(options) {
        var handlers = {};

        function mouseDown(e) {
            var keycode = e.button;

            if (isNumber(keycode) && handlers.hasOwnProperty(keycode)) {
                handlers[keycode](e);
            }
        }

        if (isObject(options)) {
            Object.keys(options).forEach(function (key) {
                var func = options[key];

                key = parseInt(key, 10);
                if (isFunction(func) && isNumber(key) && !global.isNaN(key)) {
                    handlers[key] = func;
                }
            });
            document.addEventListener('mousedown', mouseDown, false);
        }
    }

/**
 * xhrReq
 * On a successful request, the event, xhr object, and xhr response body will be passed as parameters
 * to the function assigned to options.success.  On all other events, the functions will be passed the
 * event and the xhr object. This function returns the xhr.abort method.
 * @param {Object} options
 * @return {Function} xhr.abort
 * options = {
 *     data: null,                  Data to send with the request
 *     method: 'GET',               HTTP Method to use
 *     url: global.location.href,   URL to submit the request to
 *     async: true,                 Asynchronous operations
 *     username: null,              Username credential to send with the request,
 *     password: null,              Password credential to send with the request,
 *     mimeType: null,              Response mimeType
 *     responseType: '',            "", "arraybuffer", "blob", "document", "json", and "text"
 *     headers: {},                 Sets the request headers
 *     timeout: null,               Request timeout in milliseconds
 *     success: null,               If this value is a function, it will be called upon a successful request
 *     failure: null,               If this value is a function, it will be called if the request fails
 *     readystatechange: null,      xhr.onreadystatechange event handler
 *     loadstart: null,             xhr.onloadstart event handler
 *     load: null,                  xhr.onload event handler
 *     progress: null,              xhr.onprogress event handler
 *     loadend: null,               xhr.onloadend event handler
 *     error: null,                 xhr.onerror event handler
 *     abort: null                  xhr.onabort event handler
 * };
 */
    function xhrReq(options) {
        var xhr,
            settings,
            crossOriginRegX = new RegExp(global.location.hostname),
            callback = function callback(type) {
                return function cb(e) {
                    if (type === 'readystatechange') {
                        if (xhr.readyState === 2) {
                            xhr.responseHeaders = xhr.getAllResponseHeaders();
                        } else if (xhr.readyState === 4 && (xhr.status >= 200 && xhr.status < 300)) {
                            settings.success(e, xhr, xhr.response);
                        }
                    } else if (type === 'abort' || type === 'error' || type === 'timeout') {
                        settings.failure(e, xhr);
                    }
                    if (isFunction(options[type])) {
                        options[type](e, xhr);
                    } else if (isFunction(options['on' + type])) {
                        options['on' + type](e, xhr);
                    }
                };
            };

        options = isObject(options) ? options : {};
        settings = {
            data: options.data || null,
            method: isString(options.method) ? options.method : 'GET',
            url: isString(options.url) ? options.url : global.location.href,
            async: isBoolean(options.async) ? options.async : true,
            username: isString(options.username) ? options.username : null,
            password: isString(options.password) ? options.password : null,
            mimeType: isString(options.mimeType) ? options.mimeType : null,
            responseType: isString(options.responseType) ? options.responseType : '',
            headers: isObject(options.headers) ? options.headers : {},
            timeout: isNumber(options.timeout) ? options.timeout : 0,
            success: isFunction(options.success) ? options.success : noop,
            failure: isFunction(options.failure) ? options.failure : noop
        };
        settings.crossOrigin = crossOriginRegX.test(settings.url) ? false : true;
        settings.headers['X-Requested-With'] = 'XMLHttpRequest';
        xhr = new XMLHttpRequest();
        xhr.open(settings.method, settings.url, settings.async, settings.uername, settings.password);
        Object.keys(settings.headers).forEach(function (header) {
            xhr.setRequestHeader(header, settings.headers[header]);
        });
        if (settings.crossOrigin && settings.username && settings.password) {
            xhr.withCredentials = true;
        }
        if (settings.mimeType) {
            xhr.overrideMimeType(settings.mimeType);
        }
        if (settings.responseType) {
            xhr.responseType = settings.responseType;
        }
        xhr.timeout = settings.timeout;
        xhr.onreadystatechange = callback('readystatechange');
        xhr.onloadstart = callback('loadstart');
        if (settings.method.toUpperCase() === 'POST') {
            xhr.upload.onprogress = callback('progress');
        } else {
            xhr.onprogress = callback('progress');
        }
        xhr.onload = callback('load');
        xhr.onloadend = callback('loadend');
        xhr.onerror = callback('error');
        xhr.onabort = callback('abort');
        xhr.ontimeout = callback('timeout');
        xhr.options = settings;
        xhr.send(settings.data);
        return xhr.abort.bind(xhr);
    }

/**
 * readFiles
 * On a successful request, the event, file object, the filereader, and file body will be passed as parameters
 * to the function assigned to options.success.  On all other events, the functions will be passed the
 * event, the file object, and the filereader object.
 * @param {Object} options
 * options = {
 *     element: null,       The input element with type="file" to listen on
 *     readAs: 'blob',      Read the file as: "arraybuffer", "blob", "dataurl", or "text"
 *     mimeType: '.*',      Acceptable file mimeTypes, default accepts all
 *     success: null,       If this value is a function, it will be called upon a successfully read file
 *     failure: null,       If this value is a function, it will be called if the file read files
 *     loadstart: null,     filereader.onloadstart event handler
 *     load: null,          filereader.onload event handler
 *     progress: null,      filereader.onprogress event handler
 *     loadend: null,       filereader.onloadend event handler
 *     error: null,         filereader.onerror event handler
 *     abort: null          filereader.onabort event handler
 * };
 */
    function readFiles(options) {
        var settings,
            typeMap = {
                arraybuffer: 'readAsArrayBuffer',
                blob: 'readAsArrayBuffer',
                dataurl: 'readAsDataURL',
                text: 'readAsText'
            },
            callback = function callback(type, file, filereader) {
                return function cb(e) {
                    if (type === 'error' || type === 'abort') {
                        settings.failure(e, file, filereader);
                    } else if (type === 'loadend' && (e.target.readyState === 2 || filereader.readyState === 2)) {
                        settings.success(e, file, filereader, e.target.result || filereader.result);
                    }
                    if (isFunction(options[type])) {
                        options[type](e, file, filereader);
                    } else if (isFunction(options['on' + type])) {
                        options['on' + type](e, file, filereader);
                    }
                };
            };

        function onFileSelect(e) {
            var files = e.target.files || settings.element.files,
                filereader,
                file,
                i;

            if (files && files.length) {
                for (i = 0; i < files.length; i += 1) {
                    file = files[i];
                    if (file.type.match(settings.mimeType)) {
                        filereader = new FileReader();
                        filereader.onloadstart = callback('loadstart', file, filereader);
                        filereader.onprogress = callback('progress', file, filereader);
                        filereader.onload = callback('load', file, filereader);
                        filereader.onloadend = callback('loadend', file, filereader);
                        filereader.onerror = callback('error', file, filereader);
                        filereader.onabort = callback('abort', file, filereader);
                        if (settings.readAs) {
                            filereader[settings.readAs](file);
                        }
                    }
                }
            }
        }

        options = isObject(options) ? options : {};
        settings = {
            element: isNode(options.element) ? options.element : null,
            readAs: isString(options.readAs) ? typeMap[options.readAs.toLowerCase()] : typeMap.blob,
            mimeType: isString(options.mimeType) ? options.mimeType : '.*',
            success: isFunction(options.success) ? options.success : noop,
            failure: isFunction(options.failure) ? options.failure : noop
        };
        if (settings.element && settings.element.nodeName === 'INPUT' && settings.element.type === 'file') {
            settings.element.addEventListener('change', onFileSelect, false);
        }
    }

/**
 * getPosition
 * @param {Object} options
 * options = {
 *     positionOptions: {
 *         enableHighAccuracy: false,    Enable high precision
 *         timeout: null,                Timeout in milliseconds
 *         maximumAge: null              Indicates the maximum age of a cached position
 *     },
 *     success: null,                    If this value is a function, it will be called upon successfully getting GPS coordinates
 *     failure: null                     If this value is a function, it will be called if getting GPS coordinates failed
 * };
 */
    function getPosition(options) {
        var settings;

        options = isObject(options) ? options : {};
        settings = {
            positionOptions: isObject(options.positionOptions) ? options.positionOptions : null,
            success: isFunction(options.success) ? options.success : noop,
            failure: isFunction(options.failure) ? options.failure : noop
        };
        navigator.geolocation.getCurrentPosition(settings.success, settings.failure, settings.positionOptions);
    }

/**
 * watchPosition
 * @param {Object} options
 * options = {
 *     positionOptions: {
 *         enableHighAccuracy: false,    Enable high precision
 *         timeout: null,                Timeout in milliseconds
 *         maximumAge: null              Indicates the maximum age of a cached position
 *     },
 *     success: null,                    If this value is a function, it will be called upon successfully getting GPS coordinates
 *     failure: null                     If this value is a function, it will be called if getting GPS coordinates failed
 * };
 */
    function watchPosition(options) {
        var settings;

        options = isObject(options) ? options : {};
        settings = {
            positionOptions: isObject(options.positionOptions) ? options.positionOptions : null,
            success: isFunction(options.success) ? options.success : noop,
            failure: isFunction(options.failure) ? options.failure : noop
        };
        navigator.geolocation.watchPosition(settings.success, settings.failure, settings.positionOptions);
    }

/**
 * initWebSocket
 * @param {String} url
 * @param {Object} handlers
 * handlers = {
 *     open: function () {},
 *     message: function () {},
 *     close: function () {},
 *     error: function () {}
 * };
 */
    function initWebSocket(url, handlers) {
        var socket;

        if (isFunction(global.WebSocket) && isString(url)) {
            socket = new WebSocket(url);
            if (isObject(handlers)) {
                socket.onopen = handlers.open || noop;
                socket.onmessage = handlers.message || noop;
                socket.onclose = handlers.close || noop;
                socket.onerror = handlers.error || noop;
            }
        }
        return socket
    }

/**
 * createElement
 * @param {Object} options
 * @return {Node} el
 * options = {
 *     tag: {String},
 *     attributes: {Object},
 *     styles: {Object},
 *     handlers: {Object},
 *     parent: {Node}
 * }
 */
    function createElement(options) {
        var el;

        if (isObject(options) && isString(options.tag)) {
            el = document.createElement(options.tag);
            if (el) {
                Object.keys(options.attributes).forEach(function (attr) {
                    el[attr] = options.attributes[attr];
                });
                Object.keys(options.styles).forEach(function (sty) {
                    el.style[sty] = options.styles[sty];
                });
                if (isObject(options.handlers)) {
                    Object.keys(options.handlers).forEach(function (ev) {
                        el.addEventListener(ev, options.handlers[ev], false);
                    });
                }
                if (isNode(options.parent)) {
                    options.parent.appendChild(el);
                }
            }
        }
        return el;
    }

/**
 * Cleanup
 * Return an instance of Cleanup
 * @param {String || Node || Array-Like Object}
 * @return {Object} this
 */
    function Cleanup(selector) {
        var splitSelector,
            nodes,
            x;

        if (selector instanceof Cleanup) {
            return;
        }
        if (isString(selector)) {
            selector = selector.trim();
            splitSelector = selector.split(/\s/);
            if (splitSelector.length > 1) {
                nodes = selectAll(selector);
            } else {
                if (selector.charAt(0) === '#' && !selector.match(/\./)) {
                    nodes = {
                        0: getbyid(selector.slice(1)),
                        length: 1
                    };
                } else {
                    nodes = selectAll(selector);
                }
            }
        } else if (isNode(selector)) {
            nodes = {
                0: selector,
                length: 1
            };
        } else if (isArrayLike(selector)) {
            nodes = selector;
        }
        this.length = nodes && nodes.length || 0;
        for (x = 0; x < this.length; x += 1) {
            this[x] = nodes[x];
        }
        return this;
    }

/**
 * eachnode
 * Loop through each node and pass it to the specified function.
 * @param {Function} func
 * @return {Object} this
 */
    Cleanup.prototype.eachnode = function eachnode(func) {
        return each(this, func)
    };

/**
 * attr
 * Set or get one or more attributes.
 * @param {String || Object} name
 * @param {String || Undefined} value
 * @return {Object || Array} this || values
 */
    Cleanup.prototype.attr = function attr(name, value) {
        var values;

        if (name && isString(name)) {
            name = toCamelCase(name);
            if (isString(value) || isNumber(value) || isBoolean(value)) {
                each(this, function (node) {
                    if (isObject(node)) {
                        node[name] = value;
                    }
                });
            } else {
                values = [];
                each(this, function (node) {
                    if (isObject(node)) {
                        values.push(node[name]);
                    }
                });
                return values.length === 1 ? values[0] : values;
            }
        } else if (isObject(name)) {
            Object.keys(name).forEach(function (key) {
                this.attr(key, name[key]);
            }, this);
        } else if (isArray(name)) {
            values = {};
            name.forEach(function (key) {
                values[key] = this.attr(key);
            }, this);
            return values;
        }
        return this;
    };

/**
 * remAttr
 * Remove the value of an attribute.
 * @param {String} name
 * @return {String || Array} values[0] || values
 */
    Cleanup.prototype.remAttr = function remAttr(name) {
        if (name && isString(name)) {
            name = toCamelCase(name);
            each(this, function (node) {
                if (isObject(node)) {
                    node[name] = null;
                    delete node[name];
                }
            });
        } else if (isArray(name)) {
            name.forEach(function (key) {
                this.remAttr(key);
            }, this);
        }
        return this;
    };

/**
 * prop / css
 * Set or get one or more style properties.
 * @param {String} property
 * @param {String} value
 * @return {Object} this
 */
    Cleanup.prototype.prop = function prop(name, value) {
        var values;

        if (name && isString(name)) {
            name = toCamelCase(name);
            if (isString(value) || isNumber(value) || isBoolean(value)) {
                each(this, function (node) {
                    if (isObject(node) && node.style) {
                        node.style[name] = value;
                    }
                });
            } else {
                values = [];
                each(this, function (node) {
                    if (isObject(node) && node.style) {
                        values.push(node.style[name]);
                    }
                });
                return values.length === 1 ? values[0] : values;
            }
        } else if (isObject(name)) {
            Object.keys(name).forEach(function (key) {
                this.prop(key, name[key]);
            }, this);
        } else if (isArray(name)) {
            values = {};
            name.forEach(function (key) {
                values[key] = this.prop(key);
            }, this);
            return values;
        }
        return this;
    };
    Cleanup.prototype.css = Cleanup.prototype.prop;

/**
 * remProp / remCss
 * Remove the value of a style property.
 * @param {String} prop
 * @return {String || Array} values[0] || values
 */
    Cleanup.prototype.remProp = function remProp(name) {
        if (name && isString(name)) {
            name = toCamelCase(name);
            each(this, function (node) {
                if (isObject(node) && node.style) {
                    node.style[name] = null;
                    delete node.style[name];
                }
            });
        } else if (isArray(name)) {
            name.forEach(function (key) {
                this.remProp(key);
            }, this);
        }
        return this;
    };
    Cleanup.prototype.remCss = Cleanup.prototype.remProp;

/**
 * text
 * Set or get the textContent.
 * @param {String} string
 * @return {Object} this
 */
    Cleanup.prototype.text = function text(str) {
        var values = [];

        if (isString(str) || isNumber(str)) {
            each(this, function (node) {
                if (isObject(node)) {
                    node.textContent = str;
                }
            });
        } else {
            each(this, function (node) {
                if (isObject(node)) {
                    values.push(node.textContent);
                }
            });
            return values.length === 1 ? values[0] : values;
        }
        return this;
    };

/**
 * remText
 * Remove the textContent.
 * @return {String || Array} values[0] || values
 */
    Cleanup.prototype.remText = function remText() {
        each(this, function (node) {
            if (isObject(node)) {
                node.textContent = '';
            }
        });
        return this;
    };

/**
 * html
 * Set or get the innerHTML.
 * @param {String} str
 * @return {Object} this
 */
    Cleanup.prototype.html = function html(str) {
        var values = [];

        if (isString(str) || isNumber(str)) {
            each(this, function (node) {
                if (isObject(node)) {
                    node.innerHTML = str;
                }
            });
        } else if (isUndefined(str)) {
            each(this, function (node) {
                if (isObject(node)) {
                    values.push(node.innerHTML);
                }
            });
            return values.length === 1 ? values[0] : values;
        }
        return this;
    };

/**
 * remHtml
 * Remove the innerHTML.
 * @return {String || Array} values[0] || values
 */
    Cleanup.prototype.remHtml = function remHtml() {
        each(this, function (node) {
            if (isObject(node)) {
                node.innerHTML = '';
            }
        });
        return this;
    };

/**
 * hasClass
 * Check to see if a class exists.
 * @param {String} cls
 * @return {String || Array} values[0] || values
 */
    Cleanup.prototype.hasClass = function hasClass(cls) {
        var values = [],
            rgx;

        if (cls && isString(cls)) {
            rgx = new RegExp('(^|[^-\w])(' + cls + ')([^-\w]|$)', 'g');
            each(this, function (node) {
                var test;

                if (isObject(node)) {
                    if (!isString(node.className)) {
                        node.className = '';
                    }
                    rgx.test();
                    values.push(rgx.test(node.className));
                }
            });
        }
        return values.length === 1 ? values[0] : values;
    };

/**
 * setClasses
 * Set the className, or all classes.
 * @param {String} classes
 * @return {Object} this
 */
    Cleanup.prototype.setClasses = function setClasses(classes) {
        if (isString(classes)) {
            each(this, function (node) {
                if (isObject(node)) {
                    node.className = classes.trim();
                }
            });
        }
        return this;
    };

/**
 * getClasses
 * Retrieve all classes.
 * @return {String || Array} values[0] || values
 */
    Cleanup.prototype.getClasses = function getClasses() {
        var values = [];

        each(this, function (node) {
            if (isObject(node)) {
                values.push(node.className || '');
            }
        });
        return values.length === 1 ? values[0] : values;
    };

/**
 * addClass
 * Add a class, or several separated by spaces.
 * @param {String} classes
 * @return {Object} this
 */
    Cleanup.prototype.addClass = function addClass(classes) {
        if (classes && isString(classes)) {
            classes = classes.split(' ');
            each(this, function (node) {
                if (isObject(node)) {
                    if (!isString(node.className)) {
                        node.className = '';
                    }
                    classes.forEach(function (cls) {
                        var rgx = new RegExp('(^|[^-\w])(' + cls + ')([^-\w]|$)', 'g');

                        if (!rgx.test(node.className)) {
                            node.className += ' ' + cls;
                        }
                    });
                    node.className = node.className.trim();
                }
            });
        }
        return this;
    };

/**
 * remClass
 * Remove a class, or several separated by spaces.
 * @param {String} classes
 * @return {Object} this
 */
    Cleanup.prototype.remClass = function remClass(classes) {
        if (classes && isString(classes)) {
            classes = classes.split(' ');
            each(this, function (node) {
                if (isObject(node)) {
                    if (!isString(node.className)) {
                        node.className = '';
                    }
                    classes.forEach(function (cls) {
                        var rgx = new RegExp('(^|[^-\w])(' + cls + ')([^-\w]|$)', 'g');

                        node.className = node.className.replace(rgx, '').trim();
                    });
                }
            });
        }
        return this;
    };

/**
 * togClass
 * Toggle a class or two.
 * @param {String} classone
 * @param {String} classtwo
 * @return {Object} this
 */
    Cleanup.prototype.togClass = function togClass(classone, classtwo) {
        var rgxone,
            rgxtwo;

        if (classone && isString(classone)) {
            rgxone = new RegExp('(^|[^-\w])(' + classone + ')([^-\w]|$)', 'g');
            if (classtwo && isString(classtwo) && classone !== classtwo) {
                rgxtwo = new RegExp('(^|[^-\w])(' + classtwo + ')([^-\w]|$)', 'g');
            }
            each(this, function (node) {
                if (isObject(node)) {
                    if (!isString(node.className)) {
                        node.className = '';
                    }
                    if (rgxtwo) {
                        if (rgxone.test(node.className) && !rgxtwo.test(node.className)) {
                            node.className = node.className.replace(rgxone, '').trim();
                            node.className += ' ' + classtwo;
                        } else if (!rgxone.test(node.className) && rgxtwo.test(node.className)) {
                            node.className = node.className.replace(rgxtwo, '').trim();
                            node.className += ' ' + classone;
                        } else if (rgxone.test(node.className) && rgxtwo.test(node.className)) {
                            node.className = node.className.replace(rgxtwo, '').trim();
                        } else {
                            node.className += ' ' + classone;
                        }
                    } else {
                        if (rgxone.test(node.className)) {
                            node.className = node.className.replace(rgxone, '').trim();
                        } else {
                            node.className += ' ' + classone;
                        }
                    }
                }
            });
        }
        return this;
    };


/**
 * appendTo
 * Append all nodes to a parent element.
 * @params {Node} parent
 * @return {Object} this
 */
    Cleanup.prototype.appendTo = function appendTo(parent) {
        var par = isNode(parent) ? parent : select(parent);

        if (par) {
            each(this, function (node) {
                par.appendChild(node);
            });
        }
        return this;
    };

/**
 * addEl
 * Append HTML or an element.
 * @param {String || Node} textOrNode
 * @param {String} pos
 * @return {Object} this
 */
    Cleanup.prototype.addEl = function addEl(textOrNode, pos) {
        var positions = ['beforebegin', 'afterbegin', 'beforeend', 'afterend'];

        if (isString(textOrNode) || isObject(textOrNode)) {
            each(this, function (node) {
                if (isString(textOrNode)) {
                    if (positions.indexOf(pos) === -1) {
                        pos = 'beforeend';
                    }
                    node.insertAdjacentHTML(pos, textOrNode);
                } else if (isObject(textOrNode)) {
                    node.appendChild(textOrNode);
                }
            });
        }
        return this;
    };

/**
 * remEl
 * Remove a child element.
 * @param {Node} child
 * @return {Object} this
 */
    Cleanup.prototype.remEl = function remEl(child) {
        if (isNode(child)) {
            each(this, function (node) {
                if (isNode(node) && node.contains(child)) {
                    node.removeChild(child);
                }
            });
        }
        return this;
    };

/**
 * on
 * Add an event handler.
 * @param {String} event
 * @param {Function} func
 * @param {Boolean} bub
 * @return {Object} this
 */
    Cleanup.prototype.on = function on(event, func, bub) {
        if (event && isString(event) && isFunction(func)) {
            bub = isBoolean(bub) ? bub : false;
            each(this, function (node) {
                node.addEventListener(event, func, bub);
            });
        }
        return this;
    };

/**
 * off
 * Remove an event handler.
 * @param {String} event
 * @param {Function} func
 * @param {Boolean} bub
 * @return {Object} this
 */
    Cleanup.prototype.off = function off(event, func, bub) {
        if (event && isString(event) && isFunction(func)) {
            bub = isBoolean(bub) ? bub : false;
            each(this, function (node) {
                node.removeEventListener(event, func, bub);
            });
        }
        return this;
    };

/**
 * once
 * Add an event handler that will trigger only once.
 * @param {String} event
 * @param {Function} func
 * @param {Boolean} bub
 * @return {Object} this
 */
    Cleanup.prototype.once = function once(event, func, bub) {

        function oneFunc(node) {
            return function onetime(e) {
                func.call(node, e);
                node.removeEventListener(event, onetime, bub);
            };
        }

        if (event && isString(event) && isFunction(func)) {
            bub = isBoolean(bub) ? bub : false;
            each(this, function (node) {
                node.addEventListener(event, oneFunc(node), bub);
            });
        }
        return this;
    };

    function clean(selector) {
        return new Cleanup(selector);
    }
    clean.noop = noop;
    clean.typeOf = typeOf;
    clean.isBoolean = isBoolean;
    clean.isNumber = isNumber;
    clean.isString = isString;
    clean.isArray = isArray;
    clean.isObject = isObject;
    clean.isFunction = isFunction;
    clean.isNull = isNull;
    clean.isUndefined = isUndefined;
    clean.isNode = isNode;
    clean.isElement = isElement;
    clean.isEmptyObject = isEmptyObject;
    clean.isArrayLike = isArrayLike;
    clean.arrayContains = arrayContains;
    clean.stringifiable = stringifiable;
    clean.stringify = stringify;
    clean.parsable = parsable;
    clean.parse = parse;
    clean.toCamelCase = toCamelCase;
    clean.undoCamelCase = undoCamelCase;
    clean.getbyid = getbyid;
    clean.select = select;
    clean.selectAll = selectAll;
    clean.uuid = uuid;
    clean.supplant = supplant;
    clean.inherits = inherits;
    clean.extend = extend;
    clean.each = each;
    clean.keyboardHandler = keyboardHandler;
    clean.mouseHandler = mouseHandler;
    clean.xhrReq = xhrReq;
    clean.readFiles = readFiles;
    clean.getPosition = getPosition;
    clean.watchPosition = watchPosition;
    clean.initWebSocket = initWebSocket;
    clean.createElement = createElement;

    if (!isUndefined(module) && module.hasOwnProperty('exports')) {
        module.exports = clean;
    } else {
        global.clean = clean;
    }

}(this));
