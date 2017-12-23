package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var moeDb *sql.DB

/*
Does all processing related to setting up the database for moebot
*/
func SetupDatabase(dbPass string, moeDataPass string) {
	rootDb := connectToRoot(dbPass)
	createMoebotDatabase(rootDb, moeDataPass)
	rootDb.Close()

	// actually connect with moebot now
	moeDb = openDb(createConnString("moebot", moeDataPass, "moebot"))
	createTables()
	log.Println("Finished initalizing the DB and creating tables")
}

func DisconnectAll() {
	if moeDb != nil {
		err := moeDb.Close()
		if err != nil {
			log.Println("Problem closing connection to database! - ", err)
		}
	}
}

/*
Creates all the tables for moebot
*/
func createTables() {
	// NOTE: varchar(20) for any snowflake ID's, which is the max for UINT64
	// SERVER
	moeDb.Exec(serverTable)
	// ROLE
	moeDb.Exec(roleTable)
	// CHANNEL
	moeDb.Exec(channelTable)
	// RAFFLE ENTRY
	moeDb.Exec(raffleTable)
}

/*
Creates the database and user account for moebot, if necessary
*/
func createMoebotDatabase(rootDb *sql.DB, moeDataPass string) {
	// create moebot user + database
	rows := rootDb.QueryRow("SELECT COUNT(*) FROM pg_catalog.pg_user WHERE usename = 'moebot'")
	var rowCount int
	rows.Scan(&rowCount)
	if rowCount == 0 {
		// for some reason this only works when it's non-parameterized...
		_, err := rootDb.Exec("CREATE USER moebot WITH PASSWORD '" + moeDataPass + "'")
		if err != nil {
			log.Fatal("Unable to create user moebot - ", err)
		}
	}
	rows = rootDb.QueryRow("SELECT COUNT(*) FROM pg_database WHERE datname = 'moebot'")
	rows.Scan(&rowCount)
	if rowCount == 0 {
		_, err := rootDb.Exec("CREATE DATABASE moebot OWNER moebot")
		if err != nil {
			log.Fatal("Unable to create database for moebot - ", err)
		}
	}
}

func OpenTransaction() (tx *sql.Tx) {
	tx, err := moeDb.Begin()
	if err != nil {
		log.Println("Error beginning transaction!")
		return
	}
	return
}

func connectToRoot(dbPass string) *sql.DB {
	sleepTime := 5 * time.Second
	for {
		db, err := sql.Open("postgres", createConnString("postgres", dbPass, "postgres"))
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

func createConnString(user string, pass string, db string) string {
	return "host=database user=" + user + " password=" + pass + " dbname=" + db + " sslmode=disable"
}
