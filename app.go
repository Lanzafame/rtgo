package rtgo

import (
	"code.google.com/p/go-uuid/uuid"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/chuckpreslar/emission"
	"github.com/gorilla/securecookie"
	"github.com/tpjg/goriakpbc"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type App struct {
	Emitter     *emission.Emitter
	Port        int
	Cookiename  string
	Templates   *template.Template
	Scook       *securecookie.SecureCookie
	Database    map[string]map[string]string
	Routes      map[string]map[string]string
	ConnManager map[string]*Conn
	RoomManager map[string]*Room
	DBManager   map[string]*Database
}

// ReadCookieHandler reads a secure cookie with the name specified by cookname.
// It returns the cookie value.
func (a *App) ReadCookieHandler(w http.ResponseWriter, r *http.Request) map[string]string {
	cookie, err := r.Cookie(a.Cookiename)
	if err != nil {
		return nil
	}
	cookvalue := make(map[string]string)
	if err := a.Scook.Decode(a.Cookiename, cookie.Value, &cookvalue); err != nil {
		return nil
	}
	return cookvalue
}

// SetCookieHandler sets a secure cookie with the name specified by cookname
// and with a value specified by cookvalue.
func (a *App) SetCookieHandler(w http.ResponseWriter, r *http.Request, cookname string, cookvalue map[string]string) {
	encoded, err := a.Scook.Encode(cookname, cookvalue)
	if err != nil {
		return
	}
	cookie := &http.Cookie{
		Name:  cookname,
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
	return
}

// RegisterHandler handles user registration and only parses POST requests.
func (a *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	if err := r.ParseMultipartForm(1024 * 1024); err != nil {
		w.WriteHeader(500)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	for _, db := range a.DBManager {
		if _, err := db.GetObj("users", username); err == nil {
			continue
		}
		randombytes := make([]byte, 16)
		if _, err := rand.Read(randombytes); err != nil {
			continue
		}
		salt := fmt.Sprintf("%x", sha1.Sum(randombytes))
		hashstring := []byte(fmt.Sprintf("%s%s%s%s", username, email, password, salt))
		passhash := fmt.Sprintf("%x", sha256.Sum256(hashstring))
		obj := map[string]interface{}{
			"username": username,
			"passhash": passhash,
			"email":    email,
			"salt":     salt,
			"role": map[string]interface{}{
				"privilege": "user",
				"bitmask":   0,
			},
		}
		if err := db.InsertObj("users", username, obj); err != nil {
			continue
		}
		a.SetCookieHandler(w, r, a.Cookiename, map[string]string{
			"username":  username,
			"privilege": "user",
		})
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(500)
}

// LoginHandler handles user logins and only parses POST requests.
func (a *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	if err := r.ParseMultipartForm(1024 * 1024); err != nil {
		w.WriteHeader(500)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	for _, db := range a.DBManager {
		initial, err := db.GetObj("users", username)
		if err != nil {
			continue
		}
		result := initial.(map[string]interface{})
		role := result["role"].(map[string]interface{})
		passhash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%s%s", username, result["email"], password, result["salt"])))
		if fmt.Sprintf("%x", passhash) == result["passhash"].(string) {
			a.SetCookieHandler(w, r, a.Cookiename, map[string]string{
				"username":  username,
				"privilege": role["privilege"].(string),
			})
			w.WriteHeader(200)
			return
		}
	}
	w.WriteHeader(500)
}

// BaseHandler handles the initial HTTP request and serves the base.html file.
func (a *App) BaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	cookvalue := map[string]string{
		"username":  "guest",
		"privilege": "user",
	}
	a.SetCookieHandler(w, r, a.Cookiename, cookvalue)
	a.Templates.ExecuteTemplate(w, "base", nil)
}

// StaticHandler serves all static content.
func (a *App) StaticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

// SocketHandler creates a new WebSocket connection.
func (a *App) SocketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	c := a.NewConnection(w, r)
	if c != nil {
		go c.WritePump()
		c.Join("root")
		c.ReadPump()
	}
}

// FindRoute loops through all routes attempting to match path.
// It returns the matched route.
func (a *App) FindRoute(path string) map[string]string {
	route := make(map[string]string)
	if _, ok := a.Routes[path]; ok {
		route = a.Routes[path]
	} else {
		for key, _ := range a.Routes {
			if !strings.HasPrefix(key, "^") {
				continue
			}
			reg, err := regexp.Compile(key)
			if err != nil {
				continue
			}
			match := reg.FindStringSubmatch(path)
			if match == nil || len(match) == 0 {
				continue
			}
			for k, val := range a.Routes[key] {
				if !strings.HasPrefix(val, "$") {
					route[k] = val
					continue
				}
				index, err := strconv.Atoi(string(val[1]))
				if err != nil {
					continue
				}
				route[k] = match[index]
			}
		}
	}
	return route
}

// NewConnection upgrades an icoming HTTP request, creates a new WebSocket
// connection, and adds it to ConnManager.
// It returns the new connection.
func (a *App) NewConnection(w http.ResponseWriter, r *http.Request) *Conn {
	cookie := a.ReadCookieHandler(w, r)
	socket, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	c := &Conn{
		app:       a,
		socket:    socket,
		id:        uuid.New(),
		send:      make(chan []byte, 256),
		rooms:     make(map[string]*Room),
		privilege: cookie["privilege"],
	}
	a.ConnManager[c.id] = c
	return c
}

// NewRoom will create a new room with the specified name,
// start it, and add it to RoomManager.
// It returns the new room.
func (a *App) NewRoom(name string) *Room {
	r := &Room{
		app:     a,
		name:    name,
		members: make(map[*Conn]bool),
		stop:    make(chan bool),
		join:    make(chan *Conn),
		leave:   make(chan *Conn),
		send:    make(chan []byte, 256),
	}
	go r.Start()
	a.RoomManager[name] = r
	return r
}

// NewDatabase creates a new database, adds it to DBManager, and starts it.
// It returns the new database.
func (a *App) NewDatabase(name string, params map[string]string) *Database {
	var dsn string
	var create string
	switch name {
	case "riak":
		dsn = fmt.Sprintf("%s:%s", params["host"], params["port"])
		create = ""
	case "postgres":
		dsn = fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=%s fallback_application_name=%s connect_timeout=%s sslcert=%s sslkey=%s sslrootcert=%s", params["dbname"], params["user"], params["password"], params["host"], params["sslmode"], params["fallback_application_name"], params["connect_timeout"], params["sslcert"], params["sslkey"], params["sslrootcert"])
		create = "CREATE TABLE IF NOT EXISTS %s (hash VARCHAR(255) NOT NULL UNIQUE PRIMARY KEY, data BYTEA)"
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@%s/%s?allowAllFiles=%s&allowOldPasswords=%s&charset=%s&collation=%s&clientFoundRows=%s&loc=%s&parseTime=%s&strict=%s&timeout=%s&tls=%s", params["user"], params["password"], params["host"], params["dbname"], params["allowAllFiles"], params["allowOldPasswords"], params["charset"], params["collation"], params["clientFoundRows"], params["loc"], params["parseTime"], params["strict"], params["timeout"], params["tls"])
		create = "CREATE TABLE IF NOT EXISTS %s (hash VARCHAR(255) NOT NULL UNIQUE PRIMARY KEY, data LONGBLOB)"
	case "sqlite3":
		dsn = fmt.Sprintf("%s", params["file"])
		create = "CREATE TABLE IF NOT EXISTS %s (hash VARCHAR(255) NOT NULL UNIQUE PRIMARY KEY, data BLOB)"
	}
	db := &Database{
		app:     a,
		name:    name,
		buckets: make(map[string]*riak.Bucket),
		params:  params,
		dsn:     dsn,
		create:  create,
	}
	a.DBManager[name] = db
	db.Start()
	return db
}

// Parse parses a JSON file.
func (a *App) Parse(filepath string) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Could not parse config.json: ", err)
	}
	if err := json.Unmarshal(file, a); err != nil {
		log.Fatal("Error parsing config.json: ", err)
	}
	hashKey := securecookie.GenerateRandomKey(16)
	blockKey := securecookie.GenerateRandomKey(16)
	a.Scook = securecookie.New(hashKey, blockKey)
	a.Templates = template.Must(template.ParseGlob("./static/views/*"))
}

// Start starts the app.
func (a *App) Start() {
	for dbase, params := range a.Database {
		a.NewDatabase(dbase, params)
	}
	http.HandleFunc("/", a.BaseHandler)
	http.HandleFunc("/login", a.LoginHandler)
	http.HandleFunc("/register", a.RegisterHandler)
	http.HandleFunc("/ws", a.SocketHandler)
	http.HandleFunc("/static/", a.StaticHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", a.Port), nil))
}

// NewApp parses a config.json file, instantiates the databases,
// and starts the web server.
func NewApp() *App {
	app := &App{
		Emitter:     emission.NewEmitter(),
		ConnManager: make(map[string]*Conn),
		RoomManager: make(map[string]*Room),
		DBManager:   make(map[string]*Database),
	}
	app.Parse("./config.json")
	return app
}
