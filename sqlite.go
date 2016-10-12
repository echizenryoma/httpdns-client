package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	save = *flag.Bool("db", true, "Save result to sqlite")
	path = *flag.String("path", "insight.sqlite", "Sqlite path")
)

var dbConnect *sql.DB

func initDb() {
	var err error
	if _, err = os.Stat(path); err == nil {
		dbConnect, err = sql.Open("sqlite3", path)
	} else {
		dbConnect, err = sql.Open("sqlite3", path)
		createDatabase()
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func createDatabase() {
	sql := `create table resolve (timestamp integer,time text, domain text, reply text);`
	dbConnect.Exec(sql)
}

func insertRecode(domain string, resolve string) {
	timestamp := time.Now().Unix()
	sql := fmt.Sprintf("insert into resolve (timestamp, time, domain, reply) values (%d, '%s', '%s', '%s');", timestamp, time.Now(), domain, resolve)
	dbConnect.Exec(sql)
}
