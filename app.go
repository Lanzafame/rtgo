package rtgo

import (
	"encoding/json"
	"github.com/chuckpreslar/emission"
	"github.com/gorilla/securecookie"
	"html/template"
	"io/ioutil"
	"log"
)

type App struct {
	Emitter *emission.Emitter
}

func (a *App) Start() {
	StartWebserver()
}

type RTConfig struct {
	Port       int
	Cookiename string
	Templates  *template.Template
	Scook      *securecookie.SecureCookie
	Database   map[string]map[string]string
	Routes     map[string]map[string]string
}

// config holds the values of the config.json file.
var config = &RTConfig{}

// ParseConfig parses a JSON file.
func (c *RTConfig) Parse(filepath string) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Could not parse config.json: ", err)
	}
	if err := json.Unmarshal(file, config); err != nil {
		log.Fatal("Error parsing config.json: ", err)
	}
	hashKey := securecookie.GenerateRandomKey(16)
	blockKey := securecookie.GenerateRandomKey(16)
	c.Scook = securecookie.New(hashKey, blockKey)
	c.Templates = template.Must(template.ParseGlob("./static/views/*"))
}

// NewApp parses a config.json file, instantiates the databases,
// and starts the web server.
func NewApp() *App {
	config.Parse("./config.json")
	return &App{
		Emitter: WSEmitter,
	}
}
