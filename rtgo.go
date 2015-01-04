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

type Settings struct {
	Cookiename string
	WSPath     string
	Templates  *template.Template
	HashKey    []byte
	BlockKey   []byte
	Scook      *securecookie.SecureCookie
	Database   map[string]map[string]string
	Routes     map[string]map[string]string
}

var (
	// The global config variable to which the contents of the parsed config.json file are assigned to.
	config = &Settings{}
	// The javascript line to remove from base.html when removing a controller.
	remjsline = "([[:space:]]*)<script type=\"application/javascript\" src=\"(\\.?)%s\"></script>([[:space:]]*)"
	// The javascript line to add to base.html when adding a controller.
	addjsline = "\n        <script type=\"application/javascript\" src=\"%s\"></script>\n    </body>\n"
	// The contents of a new view.
	viewtext = "{{ define \"%s\" }}\n\n{{ end }}"
	// The contents of a new controller.
	jscode = "rtgo.controllers.%s = function %s() {\n    'use strict;'\n};\n"
	// The regex to match when adding a new controller.
	bodregex = regexp.MustCompile("([[:space:]]*)</body>([[:space:]]*)")
)

// initDirectory initializes a directory.
func initDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsExist(err) {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}

// readFile reads a file and returns its contents as a string.
func readFile(filename string) (string, error) {
	fbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	file := strings.TrimSpace(string(fbytes[:]))
	return file, nil
}

// DelController deletes a controller.
func DelController(name string) error {
	basefile := "./static/views/base.html"
	file, err := readFile(basefile)
	if err != nil {
		return err
	}
	jsfilename := fmt.Sprintf("/static/js/controllers/%s.js", name)
	if strings.Contains(file, jsfilename) {
		jsregex := regexp.MustCompile(fmt.Sprintf(remjsline, jsfilename))
		newfile := jsregex.ReplaceAllString(file, "\n    ")
		if err := ioutil.WriteFile(basefile, []byte(newfile), 0644); err != nil {
			return err
		}
	}
	os.Remove("." + jsfilename)
	return nil
}

// AddController adds a controller.
func AddController(name string) error {
	initDirectory("./static/js/controllers")
	basefile := "./static/views/base.html"
	file, err := readFile(basefile)
	if err != nil {
		return err
	}
	jsfilename := fmt.Sprintf("/static/js/controllers/%s.js", name)
	if !strings.Contains(file, jsfilename) {
		jsline := fmt.Sprintf(addjsline, jsfilename)
		newfile := bodregex.ReplaceAllString(file, jsline)
		if err := ioutil.WriteFile(basefile, []byte(newfile), 0644); err != nil {
			return err
		}
	}
	if _, err := os.Stat("." + jsfilename); os.IsExist(err) {
		return nil
	}
	js := fmt.Sprintf(jscode, name, name)
	if err := ioutil.WriteFile("."+jsfilename, []byte(js), 0644); err != nil {
		return err
	}
	return nil
}

// DelView deletes a view.
func DelView(name string) error {
	viewfile := fmt.Sprintf("./static/views/%s.html", name)
	return os.Remove(viewfile)
}

// AddView adds a new view.
func AddView(name string) error {
	viewfile := fmt.Sprintf("./static/views/%s.html", name)
	if _, err := os.Stat(viewfile); os.IsExist(err) {
		return err
	}
	view := fmt.Sprintf(viewtext, name)
	if err := ioutil.WriteFile(viewfile, []byte(view), 0644); err != nil {
		return err
	}
	return nil
}

// ParseConfig parses the config.json file.
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

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 || len(args) == 1 && args[0] == "start" || args[0] == "run" {
		ParseConfig("./config.json")
		InitWebserver()
		for dbase, params := range config.Database {
			NewDatabase(dbase, params)
		}
	}
	if len(args) < 3 {
		return
	}
	if args[0] == "add" {
		if args[1] == "controller" {
			if err := AddController(args[2]); err != nil {
				log.Fatal(err)
			}
		} else if args[1] == "view" {
			if err := AddView(args[2]); err != nil {
				log.Fatal(err)
			}
		}
	} else if args[0] == "del" || args[0] == "delete" {
		if args[1] == "controller" {
			if err := DelController(args[2]); err != nil {
				log.Fatal(err)
			}
		} else if args[1] == "view" {
			if err := DelView(args[2]); err != nil {
				log.Fatal(err)
			}
		}
	}
}
