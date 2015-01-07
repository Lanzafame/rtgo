package rtgo

import (
	"encoding/json"
	"github.com/gorilla/securecookie"
	"html/template"
	"io/ioutil"
	"log"
)

type RTConfig struct {
	Port       int
	Cookiename string
	Templates  *template.Template
	HashKey    []byte
	BlockKey   []byte
	Scook      *securecookie.SecureCookie
	Database   map[string]map[string]string
	Routes     map[string]map[string]string
}

// config holds the values of the config.json file.
var config = &RTConfig{}

// ParseConfig parses a JSON file.
func ParseConfig(filepath string) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Could not parse config.json: ", err)
	}
	if err := json.Unmarshal(file, config); err != nil {
		log.Fatal("Error parsing config.json: ", err)
	}
	config.Templates = template.Must(template.ParseGlob("./static/views/*"))
	config.HashKey = securecookie.GenerateRandomKey(16)
	config.BlockKey = securecookie.GenerateRandomKey(16)
	config.Scook = securecookie.New(config.HashKey, config.BlockKey)
}

// NewApp parses a config.json file, instantiates the databases,
// and starts the web server.
func NewApp() {
	ParseConfig("./config.json")
	for dbase, params := range config.Database {
		NewDatabase(dbase, params)
	}
	StartWebserver()
}
