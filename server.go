package rtgo

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
)

// ReadCookieHandler reads a secure cookie with the name specified by cookname.
// It returns the cookie value.
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

// SetCookieHandler sets a secure cookie with the name specified by cookname
// and with a value specified by cookvalue.
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

// RegisterHandler handles user registration and only parses POST requests.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
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
	for _, db := range DBManager {
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
		SetCookieHandler(w, r, config.Cookiename, map[string]string{
			"username":  username,
			"privilege": "user",
		})
		w.WriteHeader(200)
		return
	}
	w.WriteHeader(500)
}

// LoginHandler handles user logins and only parses POST requests.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

// BaseHandler handles the initial HTTP request and serves the base.html file.
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

// StaticHandler serves all static content.
func StaticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

// SocketHandler creates a new WebSocket connection.
func SocketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	c := NewConnection(w, r)
	if c != nil {
		go c.WritePump()
		c.Join("root")
		c.ReadPump()
	}
}

// StartWebserver starts the web server.
func StartWebserver() {
	http.HandleFunc("/", BaseHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/ws", SocketHandler)
	http.HandleFunc("/static/", StaticHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
