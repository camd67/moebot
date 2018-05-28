package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type MoeWebDb struct {
	rootPassword string
	moeUser      string
	moePassword  string
	host         string
	db           *sql.DB
	tables       []dataTable
	Users        *userTable
	GuildEvents  *guildEventTable
}

type dataTable interface {
	createTable()
}

func NewDatabase(host string, rootPassword string, user string, password string) *MoeWebDb {
	return &MoeWebDb{host: host, rootPassword: rootPassword, moeUser: user, moePassword: password}
}

func (d *MoeWebDb) Initialize() error {
	err := d.createDb()
	if err != nil {
		return err
	}
	d.db = openDb(createConnString(d.host, d.moeUser, d.moePassword, "moebot_www"))
	d.db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	d.createTables()
	log.Println("Finished initalizing the DB and creating tables")
	return nil
}

func (d *MoeWebDb) DisconnectAll() {
	if d.db != nil {
		err := d.db.Close()
		if err != nil {
			log.Println("Problem closing connection to database! - ", err)
		}
	}
}

func (d *MoeWebDb) createDb() error {
	var err error
	rootDb := d.connectToRoot()
	// create moebot user + database
	rows := rootDb.QueryRow("SELECT COUNT(*) FROM pg_catalog.pg_user WHERE usename = $1", d.moeUser)
	var rowCount int
	rows.Scan(&rowCount)
	if rowCount == 0 {
		// for some reason this only works when it's non-parameterized...
		_, err = rootDb.Exec("CREATE USER " + d.moeUser + " WITH PASSWORD '" + d.moePassword + "'")
		if err != nil {
			log.Fatal("Unable to create main user - ", err)
		}
	}
	rows = rootDb.QueryRow("SELECT COUNT(*) FROM pg_database WHERE datname = 'moebot_www'")
	rows.Scan(&rowCount)
	if rowCount == 0 {
		_, err = rootDb.Exec("CREATE DATABASE moebot_www OWNER " + d.moeUser)
		if err != nil {
			log.Fatal("Unable to create database for moebot_www - ", err)
		}
	}
	rootDb.Close()
	return err
}

func (d *MoeWebDb) connectToRoot() *sql.DB {
	sleepTime := 5 * time.Second
	for {
		db, err := sql.Open("postgres", createConnString(d.host, "postgres", d.rootPassword, "postgres"))
		if err != nil {
			log.Println("Unable to open DB connection", err)
			log.Println("Waiting before attempting to reconnect")
			time.Sleep(sleepTime)
			continue
		}
		err = db.Ping()
		if err != nil {
			log.Println("Unable to ping DB", err)
			log.Println("Waiting before attempting to reconnect")
			time.Sleep(sleepTime)
			continue
		}
		// keep looping till we get past all the error checks
		return db
	}
}

func (d *MoeWebDb) createTables() {
	d.Users = &userTable{d.db}
	d.GuildEvents = &guildEventTable{d.db}
	d.tables = []dataTable{d.Users, d.GuildEvents}
	for _, t := range d.tables {
		t.createTable()
	}
}

func openDb(connString string) *sql.DB {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatal("Unable to connect to database - ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging newly connected db")
	}
	return db
}

func createConnString(host string, user string, pass string, db string) string {
	return "host=" + host + " user=" + user + " password=" + pass + " dbname=" + db + " sslmode=disable"
}
