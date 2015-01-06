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


## API
- **type RTConfig struct**
  - **Port int** - the port number the HTTP server will listen on.
  - **Cookiename string** - the name of the cookie that will be assigned to each user.
  - **Templates *template.Template** - stores the views.
  - **HashKey []byte** - the hash key used when creating a secure cookie.
  - **BlockKey []byte** - the block key used when creating a secure cookie.
  - **Scook *securecookie.SecureCookie** - the secure cookie to use when assigning a cookie to the user.
  - **Database map[string]string** - a map of the databases specified in the config.json file.
  - **Routes map[string]string** - a map of the routes specified in the config.json file.
- **type RTConn struct**
  - **socket *ws.Conn** - the websocket connection interface
  - **id uuid.New()** - a unique identifier identifying this connection
  - **send chan []byte** - the channel by which messages are sent and received
  - **rooms map[string]*RTRoom** - a map of rooms this connection is in
  - **privilege string** - this will be either "user" or "admin"; by default it is set to "user"
  - **func (c *RTConn) ReadPump**
  - **func (c *RTConn) Write(mt int, payload []byte) error**
  - **func (c *RTConn) WritePump()**
  - **func (c *RTConn) SendView(path string)** - send the view labeled in config.json under the path specified.
  - **func (c *RTConn) Join(name string)** - join a room
  - **func (c *RTConn) Leave(name string)** - leave a room
  - **func (c *RTConn) Emit(payload *Message)** - send a message to all connection in the room specified in the payload
- **type RTRoom struct**
  - **func (r *RTRoom) Start()**
  - **func (r *RTRoom) Stop()**
  - **func (r *RTRoom) Join(c *RTConn)**
  - **func (r *RTRoom) Leave(c *RTConn)**
  - **func (r *RTRoom) Emit(payload *Message)**
- **type RTDatabase struct**
  - **func (db *RTDatabase) GetAllObjs(table string) ([]interface{}, error)**
  - **func (db *RTDatabase) GetObj(table string, key string) (interface{}, error)**
  - **func (db *RTDatabase) DeleteObj(table string, key string) error**
  - **func (db *RTDatabase) InsertObj(table string, key string, data interface{}) error**
  - **func (db *RTDatabase) Start()**
- **type Message struct**
  - **Room string `json:"room"`**
  - **Event string `json:"event"`**
  - **Payload string `json:"payload"`**
- **type DBMessage struct**
  - **DB string `json:"db"`**
  - **Table string `json:"table"`**
  - **Key string `json:"key"`**
  - **Data string `json:"data"`**
- **ConnManager map[string]*RTConn**
- **RoomManager map[string]*RTRoom**
- **DBManager map[string]*RTDatabase**
- **func NewApp()**
- **func NewConnection(w http.ResponseWriter, r *http.Request) *RTConn**
- **func NewRoom(name string) *RTRoom**
- **func NewDatabase(name string, params map[string]string) *RTDatabase**
- **func ReadCookieHandler(w http.ResponseWriter, r *http.Request, cookname string) map[string]string**
- **func SetCookieHandler(w http.ResponseWriter, r *http.Request, cookname string, cookvalue map[string]string)**
- **func RegisterHandler(w htp.ResponseWriter, r *http.Request)**
- **func LoginHandler(w http.ResponseWriter, r *http.Request)**
- **func BaseHandler(w http.ResponseWriter, r *http.Request)**
- **func StaticHandler(w http.ResponseWriter, r *http.Request)**
- **func SocketHandler(w http.ResponseWriter, r *http.Request)**
- **func StartWebserver()**
- **func HandleData(c *RTConn, data *Message) error* - handles all incoming parsed JSON blobs
