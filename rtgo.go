package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/securecookie"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type Config struct {
	Templates *template.Template
	HashKey   []byte
	BlockKey  []byte
	Scook     *securecookie.SecureCookie
	Database  map[string]map[string]string
	Routes    map[string]map[string]string
}

var (
	config      Config
	ConnManager = make(map[string]*RTConn)
	RoomManager = make(map[string]*RTRoom)
	DBManager   = make(map[string]*RTDatabase)
)

func initDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsExist(err) {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}

func readFile(filename string) (string, error) {
	fbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	file := strings.TrimSpace(string(fbytes[:]))
	return file, nil
}

func DelController(name string) error {
	basefile := "./static/views/base.html"
	file, err := readFile(basefile)
	if err != nil {
		return err
	}
	jsfilename := fmt.Sprintf("/static/js/controllers/%s.js", name)
	if strings.Contains(file, jsfilename) {
		jsregex := regexp.MustCompile(fmt.Sprintf("([[:space:]]*)<script type=\"application/javascript\" src=\"(\\.?)%s\"></script>([[:space:]]*)", jsfilename))
		newfile := jsregex.ReplaceAllString(file, "\n    ")
		if err := ioutil.WriteFile(basefile, []byte(newfile), 0644); err != nil {
			return err
		}
	}
	os.Remove("." + jsfilename)
	return nil
}

func AddController(name string) error {
	initDirectory("./static/js/controllers")
	basefile := "./static/views/base.html"
	file, err := readFile(basefile)
	if err != nil {
		return err
	}
	jsfilename := fmt.Sprintf("/static/js/controllers/%s.js", name)
	if !strings.Contains(file, jsfilename) {
		jsregex := regexp.MustCompile("([[:space:]]*)</body>([[:space:]]*)")
		jsline := fmt.Sprintf("\n        <script type=\"application/javascript\" src=\"%s\"></script>\n    </body>\n", jsfilename)
		newfile := jsregex.ReplaceAllString(file, jsline)
		if err := ioutil.WriteFile(basefile, []byte(newfile), 0644); err != nil {
			return err
		}
	}
	if _, err := os.Stat("." + jsfilename); os.IsExist(err) {
		return nil
	}
	js := fmt.Sprintf("rtgo.controllers.%s = function %s() {\n    'use strict;'\n};\n", name, name)
	if err := ioutil.WriteFile("."+jsfilename, []byte(js), 0644); err != nil {
		return err
	}
	return nil
}

func ParseConfig(filepath string) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal("Could not parse config.json: ", err)
	}
	if err := json.Unmarshal(file, &config); err != nil {
		log.Fatal("Error parsing config.json: ", err)
	}
	config.Templates = template.Must(template.ParseGlob("./static/views/*"))
	config.HashKey = securecookie.GenerateRandomKey(16)
	config.BlockKey = securecookie.GenerateRandomKey(16)
	config.Scook = securecookie.New(config.HashKey, config.BlockKey)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 1 && args[0] == "start" {
		ParseConfig("./config.json")
		InitServer()
	}
	if len(args) < 3 {
		return
	}
	if args[0] == "add" {
		if args[1] == "controller" {
			err := AddController(args[2])
			if err != nil {
				log.Fatal(err)
			}
		} else if args[1] == "view" {

		}
	} else if args[0] == "del" || args[0] == "delete" {
		if args[1] == "controller" {
			err := DelController(args[2])
			if err != nil {
				log.Fatal(err)
			}
		} else if args[1] == "view" {

		}
	}
}
