package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/net/publicsuffix"
)

var (
	save = *flag.Bool("save", true, "Whether to save the results to a sqlite")
)

var dbConnect *sql.DB
var dbPath = fmt.Sprintf("%d-%02d.sqlite", time.Now().Year(), time.Now().Month())

func initDb() {
	log.Printf("Resolve result save at %s", dbPath)
	var err error
	if _, err = os.Stat(dbPath); err == nil {
		dbConnect, err = sql.Open("sqlite3", dbPath)
	} else {
		dbConnect, err = sql.Open("sqlite3", dbPath)
		createDatabase()
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func createDatabase() {
	sql := `create table resolve (timestamp integer, time text, tld text, domain text, ips text);`
	dbConnect.Exec(sql)
}

func insertRecode(domain string, ips string) {
	now := time.Now()
	domain = domain[0 : len(domain)-1]
	tld, _ := publicsuffix.EffectiveTLDPlusOne(domain)
	_, err := dbConnect.Exec("insert into resolve (timestamp, time, tld, domain, ips) values (?, ?, ?, ?, ?)", now.Unix(), now, tld, domain, ips)
	if err != nil {
		log.Println(err.Error())
	}
}
