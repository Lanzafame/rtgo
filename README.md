rtgo
====

A Go real-time web framework that all starts with a config.json file. 

## config.json
There is an example config.json file (config.json.example) in the repo which clearly depicts the possible fields.  I have specified them below as well:
- **port** - the port for which the HTTP server will listen on
- **cookiename** - the name of the cookie to be used
- **database** - an object specifying the databases to use
  - **postgres** - http://godoc.org/github.com/lib/pq
  - **mysql** - https://github.com/go-sql-driver/mysql
  - **sqlite3** - http://godoc.org/github.com/mattn/go-sqlite3
  - **riak** - https://github.com/tpjg/goriakpbc
- **routes**
  - **route** - route can be either a string or a regular expression
    - **table** - the name of the database table to query upon the request for this route
    - **key** - the key to query from the table above, and whose value will be rendered into the template specified below; if no key is specified, all values will be gotten
    - **template** - the template to render when this route is requested; the database values in the table specified above will be rendered within the template
    - **controller** - the javascript controller associated with and run when this route is requested, and the template is rendered


## DOM
- **data-rt-view=""** - Assign this attribute to the element which will act as the container for requested views. By default, this is already specified in base.html.
- **data-rt-href="{path}"** - All elements with this attribute will have on onclick listener attached to them. When clicked, the corresponding view will be requested.


## rtgo.js
- **rtgo.showLogin()** - show the login form
- **rtgo.showRegister()** - show the register form
- **rtgo.hideForms()** - hide all visible forms
By default the below functions will not go through unless the user calling them is an admin.
- **rtgo.getObj(db, table, key)** - get an object from a database
- **rtgo.insertObj(db, table, key, data)** - insert an object into a database
- **rtgo.deleteObj(db, table, key)** - delete an object from a database

## command-line tool
Run `go build -o $GOBIN/rtgo cmd/rtgo.go`

This is a simple command-line tool where the following commands can be run:
- **rtgo add controller &lt;name&gt;**
- **rtgo del controller &lt;name&gt;**
- **rtgo add view &lt;name&gt;**
- **rtgo add view &lt;name&gt;**


## Example
```go
package main

import "github.com/jdeezy/rtgo"

var app *rtgo.App

func main() {
    app = rtgo.NewApp()
    app.Emitter.On("event-name", func(conn *rtgo.Conn, data *rtgo.Message) {
        // do something here
    })
    app.Start()
}
```
