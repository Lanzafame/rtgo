package rtgo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tpjg/goriakpbc"
	"log"
	"strings"
)

type RTDatabase struct {
	name       string
	params     map[string]string
	buckets    map[string]*riak.Bucket
	dsn        string
	create     string
	connection *sql.DB
}

var DBManager = make(map[string]*RTDatabase)

// GetAllObjs returns an array of all objects in a database table.
func (db *RTDatabase) GetAllObjs(table string) ([]interface{}, error) {
	data := make([]interface{}, 0)
	if db.name == "riak" {
		if _, exists := db.buckets[table]; !exists {
			return nil, errors.New("Bucket does not exist.")
		}
		keys, err := db.buckets[table].ListKeys()
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			collect := make(map[string]string)
			obj, err := db.GetObj(table, string(key))
			if err != nil {
				return nil, err
			}
			collect["hash"] = string(key)
			collect["data"] = obj.(string)
			data = append(data, collect)
		}
	} else {
		query := fmt.Sprintf("SELECT * FROM %s", table)
		rows, err := db.connection.Query(query)
		if err != nil {
			return nil, err
		}
		cols, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		blobs := make([][]byte, len(cols))
		dest := make([]interface{}, len(cols))
		for i := range cols {
			dest[i] = &blobs[i]
		}
		for rows.Next() {
			err := rows.Scan(dest...)
			if err != nil {
				return nil, err
			}
			collect := make(map[string]string)
			for i, blob := range blobs {
				var value interface{}
				col := cols[i]
				if col == "hash" {
					value = string(blob)
				} else if err := json.Unmarshal(blob, &value); err != nil {
					return nil, err
				}
				collect[col] = value.(string)
			}
			data = append(data, collect)
		}
	}
	return data, nil
}

// GetObj retrieves an object from a database table.
func (db *RTDatabase) GetObj(table string, key string) (interface{}, error) {
	var data interface{}
	if db.name == "riak" {
		if _, exists := db.buckets[table]; !exists {
			return nil, errors.New("Bucket does not exist.")
		}
		if exists, _ := db.buckets[table].Exists(key); !exists {
			return nil, errors.New("Object does not exist.")
		}
		obj, err := db.buckets[table].Get(key)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(obj.Data, &data); err != nil {
			return nil, err
		}
	} else {
		query := ""
		blob := make([]byte, 0)
		if db.name == "postgres" {
			query = fmt.Sprintf("SELECT data FROM %s WHERE hash = $1", table)
		} else {
			query = fmt.Sprintf("SELECT data FROM %s WHERE hash = ?", table)
		}
		if err := db.connection.QueryRow(query, key).Scan(&blob); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(blob, &data); err != nil {
			return nil, err
		}
	}
	return data, nil
}

// DeleteObj deletes an object from a database table.
func (db *RTDatabase) DeleteObj(table string, key string) error {
	if db.name == "riak" {
		if _, exists := db.buckets[table]; !exists {
			return errors.New("Bucket does not exist.")
		}
		if err := db.buckets[table].Delete(key); err != nil {
			return err
		}
	} else {
		query := ""
		if db.name == "postgres" {
			query = fmt.Sprintf("DELETE FROM %s WHERE hash = $1", table)
		} else {
			query = fmt.Sprintf("DELETE FROM %s WHERE hash = ?", table)
		}
		if _, err := db.connection.Exec(query, key); err != nil {
			return err
		}
	}
	return nil
}

// InsertObj inserts an object into a database table.
func (db *RTDatabase) InsertObj(table string, key string, data interface{}) error {
	blob, err := json.Marshal(&data)
	if err != nil {
		return err
	}
	if db.name == "riak" {
		if _, exists := db.buckets[table]; !exists {
			return errors.New("Bucket does not exist.")
		}
		obj := db.buckets[table].NewObject(key)
		obj.ContentType = "application/json"
		obj.Data = blob
		if err = obj.Store(); err != nil {
			return err
		}
	} else {
		query := ""
		if db.name == "postgres" {
			query = fmt.Sprintf("INSERT INTO %s (hash, data) VALUES ($1, $2)", table)
		} else {
			query = fmt.Sprintf("INSERT INTO %s (hash, data) VALUES (?, ?)", table)
		}
		if _, err := db.connection.Exec(query, key, blob); err != nil {
			return err
		}
	}
	return nil
}

// Start starts the database and initializes the database tables/buckets.
// If a users table is not specified in the config.json file, create it anyways.
func (db *RTDatabase) Start() {
	usersTableExists := false
	if db.name == "riak" {
		if err := riak.ConnectClient(db.dsn); err != nil {
			log.Fatal("Cannot connect, is Riak running?")
		}
		tableList := strings.Split(db.params["buckets"], ",")
		for _, bname := range tableList {
			if bname == "users" {
				usersTableExists = true
			}
			db.buckets[bname], _ = riak.NewBucket(bname)
		}
		if usersTableExists == false {
			db.buckets["users"], _ = riak.NewBucket("users")
		}
	} else {
		dbconn, err := sql.Open(db.name, db.dsn)
		if err != nil {
			log.Fatal(err)
		}
		db.connection = dbconn
		if _, exists := db.params["tables"]; !exists {
			return
		}
		tableList := strings.Split(db.params["tables"], ",")
		for _, table := range tableList {
			if table == "users" {
				usersTableExists = true
			}
			statement := fmt.Sprintf(db.create, table)
			if _, err := db.connection.Exec(statement); err != nil {
				log.Fatal(err)
			}
		}
		if usersTableExists == false {
			statement := fmt.Sprintf(db.create, "users")
			if _, err := db.connection.Exec(statement); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// NewDatabase returns a new instance of RTDatabase
func NewDatabase(name string, params map[string]string) *RTDatabase {
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
	db := &RTDatabase{
		name:    name,
		buckets: make(map[string]*riak.Bucket),
		params:  params,
		dsn:     dsn,
		create:  create,
	}
	// Add the new instance of RTDatabase to the DBManager
	DBManager[name] = db
	db.Start()
	return db
}
