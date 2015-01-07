rtgo
====

A Go real-time web framework that all starts with a config.json file.  Right now this is alpha version.


## config.json
There is an example config.json file (config.json.example) in the repo which clearly depicts the possible fields.  I have specified them below as well:
- **port** - the port 
- **cookiename** - the name of the cookie to be used
- **database** - an object specifying the databases to use
  - **postgres** - http://godoc.org/github.com/lib/pq
  - **mysql** - https://github.com/go-sql-driver/mysql
  - **sqlite3** - http://godoc.org/github.com/mattn/go-sqlite3
  - **riak** - https://github.com/tpjg/goriakpbc
- **routes**
  - **route** - route can be either a string or a regular expression
    - **table** - the name of the database table to query upon the request for this route
    - **template** - the template to render when this route is requested; the database values in the table specified above will be rendered within the template
    - **controller** - the javascript controller associated with and run when this route is requested, and the template is rendered


## DOM
- **data-rt-view=""** - Assign this attribute to the element which will act as the container for requested views. By default, this is already specified in base.html.
- **data-rt-href="{path}"** - All elements with this attribute will have on onclick listener attached to them. When clicked, the corresponding view will be requested.
