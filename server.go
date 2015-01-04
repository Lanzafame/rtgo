package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Read a secure cookie and return its value.
func ReadCookieHandler(w http.ResponseWriter, r *http.Request, cookname string) map[string]string {
	cookie, err := r.Cookie(cookname)
	if err != nil {
		return nil
	}
	cookvalue := make(map[string]string)
	if err := config.Scook.Decode(cookname, cookie.Value, &cookvalue); err != nil {
		return nil
	}
	return cookvalue
}

// Set a secure cookie with a cookie name and value.
func SetCookieHandler(w http.ResponseWriter, r *http.Request, cookname string, cookvalue map[string]string) {
	encoded, err := config.Scook.Encode(cookname, cookvalue)
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

// RegisterHandler handles user registration and only handles POST requests.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	// Parse the incoming form data.
	if err := r.ParseMultipartForm(1024 * 1024); err != nil {
		w.WriteHeader(500)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	// Loop through the list of database connections held in DBManager.
	for _, db := range DBManager {
		// If the specified username already exists, move on to the next database.
		if _, err := db.GetObj("users", username); err == nil {
			continue
		}
		randombytes := make([]byte, 16)
		if _, err := rand.Read(randombytes); err != nil {
			continue
		}
		// Generate a SHA256 hash for the parsed user credentials.
		// SHA256Sum(username + email + password + salt)
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
		// Insert the new user credentials into the database.
		if err := db.InsertObj("users", username, obj); err != nil {
			continue
		}
		// Set a secure cookie with the cookiename specified in config.json if the
		// new user credentials were successfully entered into the database.
		SetCookieHandler(w, r, config.Cookiename, map[string]string{
			"username":  username,
			"privilege": "user",
		})
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(500)
}

// LoginHandler handles user logins and only handles POST requests.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", 405)
		return
	}
	// Parse the received form data.
	if err := r.ParseMultipartForm(1024 * 1024); err != nil {
		w.WriteHeader(500)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	for _, db := range DBManager {
		initial, err := db.GetObj("users", username)
		if err != nil {
			continue
		}
		result := initial.(map[string]interface{})
		role := result["role"].(map[string]interface{})
		passhash := sha256.Sum256([]byte(fmt.Sprintf("%s%s%s%s", username, result["email"], password, result["salt"])))
		if fmt.Sprintf("%x", passhash) == result["passhash"].(string) {
			SetCookieHandler(w, r, config.Cookiename, map[string]string{
				"username":  username,
				"privilege": role["privilege"].(string),
			})
			w.WriteHeader(200)
			return
		}
	}
	w.WriteHeader(500)
}

// BaseHandler server the base.html file and handles the initial request
func BaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	cookvalue := map[string]string{
		"username":  "guest",
		"privilege": "user",
	}
	SetCookieHandler(w, r, config.Cookiename, cookvalue)
	config.Templates.ExecuteTemplate(w, "base", nil)
}

// StaticHandler serves the static content
func StaticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

// FindRoute returns the variables specified in the config.json
// file that match the path requested.
func FindRoute(path string) map[string]string {
	route := make(map[string]string)
	if _, ok := config.Routes[path]; ok {
		route = config.Routes[path]
	} else {
		for key, _ := range config.Routes {
			reg, err := regexp.Compile(key)
			if err != nil {
				continue
			}
			match := reg.FindStringSubmatch(path)
			if match == nil || len(match) == 0 {
				continue
			}
			for k, val := range config.Routes[key] {
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

// SendView sends the view associated with the requested path.
func SendView(conn *RTConn, path string) {
	var doc bytes.Buffer
	var err error
	route := FindRoute(path)
	if _, ok := route["template"]; !ok {
		log.Println("No template for the specified path: ", path)
		return
	}
	collection := make([]interface{}, 0)
	// If a table is specified in the config.json file under the matched
	// route, retrieve the values in the table.
	if _, ok := route["table"]; ok {
		for _, db := range DBManager {
			collection, err = db.GetAllObjs(route["table"])
			if err != nil {
				continue
			}
			break
		}
	}
	// Render the retrieved database values in the template specified in the
	// config.json file for the requested route.
	config.Templates.ExecuteTemplate(&doc, route["template"], collection)
	response := map[string]interface{}{
		"room":  "root",
		"event": "response",
		"payload": map[string]string{
			"template":   doc.String(),
			"controller": route["controller"],
		},
	}
	data, err := json.Marshal(&response)
	if err != nil {
		log.Println("error encoding json: ", err)
		return
	}
	conn.send <- data
}

// InitWebserver starts the webserver.
func InitWebserver() {
	http.HandleFunc("/", BaseHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/ws", SocketHandler)
	http.HandleFunc("/static/", StaticHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
