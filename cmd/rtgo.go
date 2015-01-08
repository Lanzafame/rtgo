package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	// The javascript line to remove from base.html when removing a controller.
	remjsline = "([[:space:]]*)<script type=\"application/javascript\" src=\"(\\.?)%s\"></script>([[:space:]]*)"
	// The javascript line to add to base.html when adding a controller.
	addjsline = "\n        <script type=\"application/javascript\" src=\"%s\"></script>\n    </body>\n"
	// The contents of a new view.
	viewtext = "{{ define \"%s\" }}\n\n{{ end }}"
	// The contents of a new controller.
	jscode = "rtgo.controllers.%s = function %s() {\n    'use strict;'\n};\n"
	// The regex to match when adding a new controller.
	bodregex   = regexp.MustCompile("([[:space:]]*)</body>([[:space:]]*)")
	run        = flag.Bool("run", false, "Run the RTGo server.")
	add        = flag.Bool("add", false, "Add either a view or controller.")
	del        = flag.Bool("del", false, "Delete either a view or controller.")
	view       = flag.String("view", "", "The name of the view to add or delete.")
	controller = flag.String("controller", "", "The name of the controller to add or delete.")
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

func main() {
	flag.Parse()
	if *add {
		if *controller != "" {
			if err := AddController(*controller); err != nil {
				log.Fatal(err)
			}
		}
		if *view != "" {
			if err := AddView(*view); err != nil {
				log.Fatal(err)
			}
		}
	} else if *del {
		if *controller != "" {
			if err := DelController(*controller); err != nil {
				log.Fatal(err)
			}
		}
		if *view != "" {
			if err := DelView(*view); err != nil {
				log.Fatal(err)
			}
		}
	}
}
