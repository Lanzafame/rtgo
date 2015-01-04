package main

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
	parameters map[string]string
	buckets    map[string]*riak.Bucket
	dsn        string
	create     string
	connection *sql.DB
}

// retrieve all objects/data stored within a specified database table
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

// get an object from a table
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

// delete an object from a table
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

// insert an object into a table
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

func (db *RTDatabase) Start() {
	usersTableExists := false
	if db.name == "riak" {
		if err := riak.ConnectClient(db.dsn); err != nil {
			log.Fatal("Cannot connect, is Riak running?")
		}
		tableList := strings.Split(db.parameters["buckets"], ",")
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
		if _, exists := db.parameters["tables"]; !exists {
			return
		}
		tableList := strings.Split(db.parameters["tables"], ",")
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

func NewDatabase(name string, parameters map[string]string) *RTDatabase {
	var dsn string
	var create string
	switch name {
	case "riak":
		dsn = fmt.Sprintf("%s:%s", parameters["host"], parameters["port"])
		create = ""
	case "postgres":
		dsn = fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=%s fallback_application_name=%s connect_timeout=%s sslcert=%s sslkey=%s sslrootcert=%s", parameters["dbname"], parameters["user"], parameters["password"], parameters["host"], parameters["sslmode"], parameters["fallback_application_name"], parameters["connect_timeout"], parameters["sslcert"], parameters["sslkey"], parameters["sslrootcert"])
		create = "CREATE TABLE IF NOT EXISTS %s (hash VARCHAR(255) NOT NULL UNIQUE PRIMARY KEY, data BYTEA)"
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@%s/%s?allowAllFiles=%s&allowOldPasswords=%s&charset=%s&collation=%s&clientFoundRows=%s&loc=%s&parseTime=%s&strict=%s&timeout=%s&tls=%s", parameters["user"], parameters["password"], parameters["host"], parameters["dbname"], parameters["allowAllFiles"], parameters["allowOldPasswords"], parameters["charset"], parameters["collation"], parameters["clientFoundRows"], parameters["loc"], parameters["parseTime"], parameters["strict"], parameters["timeout"], parameters["tls"])
		create = "CREATE TABLE IF NOT EXISTS %s (hash VARCHAR(255) NOT NULL UNIQUE PRIMARY KEY, data LONGBLOB)"
	case "sqlite3":
		dsn = fmt.Sprintf("%s", parameters["file"])
		create = "CREATE TABLE IF NOT EXISTS %s (hash VARCHAR(255) NOT NULL UNIQUE PRIMARY KEY, data BLOB)"
	}
	database := &RTDatabase{
		name:       name,
		buckets:    make(map[string]*riak.Bucket),
		parameters: parameters,
		dsn:        dsn,
		create:     create,
	}
	DBManager[name] = database
	return database
}

// start the database
func InitDatabases() {
	for dbase, parameters := range config.Database {
		database := NewDatabase(dbase, parameters)
		database.Start()
	}
}
