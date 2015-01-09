//    Title: db.go
//    Author: JD
//
//    This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.

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

type Database struct {
	app        *App
	name       string
	params     map[string]string
	buckets    map[string]*riak.Bucket
	dsn        string
	create     string
	connection *sql.DB
}

// GetAllObjs selects all rows and columns in a database table.
// It returns an array of interfaces or an error.
func (db *Database) GetAllObjs(table string) ([]interface{}, error) {
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
			collect := make(map[string]interface{})
			obj, err := db.GetObj(table, string(key))
			if err != nil {
				return nil, err
			}
			collect["hash"] = string(key)
			collect["data"] = obj
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
			collect := make(map[string]interface{})
			for i, blob := range blobs {
				var value interface{}
				col := cols[i]
				if col == "hash" {
					value = string(blob)
				} else if err := json.Unmarshal(blob, &value); err != nil {
					return nil, err
				}
				collect[col] = value
			}
			data = append(data, collect)
		}
	}
	return data, nil
}

// GetObj selects data from a table with the matching key.
// It returns an interface or an error.
func (db *Database) GetObj(table string, key string) (interface{}, error) {
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

// DeleteObj deletes a row from a database table with a matching key.
// It may return an error.
func (db *Database) DeleteObj(table string, key string) error {
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

// InsertObj inserts data into a database table with the specified key.
// It may return an error.
func (db *Database) InsertObj(table string, key string, data interface{}) error {
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

// Start starts the database and initializes its tables/buckets.
// If a users table is not specified in the config.json file,
// one is created anyways.
func (db *Database) Start() {
	usersTableExists := false
	if db.name == "riak" {
		if err := riak.ConnectClient(db.dsn); err != nil {
			log.Fatal("Cannot connect, is Riak running?")
		}
		tableList := strings.Split(db.params["tables"], ",")
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
