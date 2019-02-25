/*
All database interactions and database tables within moebot's database
*/
package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

var moeDb *sql.DB

const (
	DbMaxUidLength         = 20
	MaxMessageLength       = 1900
	MaxMessageLengthString = "1900"
)

/*
Does all processing related to setting up the database for moebot
*/
func SetupDatabase(host string, moeDataPass string) {
	moeDb = openDb(createConnString(host, "moebot", moeDataPass, "moebot"))
	log.Println("Finished initializing the DB and creating tables")
}

func DisconnectAll() {
	if moeDb != nil {
		err := moeDb.Close()
		if err != nil {
			log.Println("Problem closing connection to database! - ", err)
		}
	}
}

func openDb(connString string) *sql.DB {
	sleepTime := 5 * time.Second
	for {
		db, err := sql.Open("postgres", connString)
		if err != nil {
			log.Println("Unable to connect to database", err)
			log.Println("Waiting before attempting to reconnect")
			time.Sleep(sleepTime)
			continue
		}

		err = db.Ping()
		if err != nil {
			err = db.Close()
			if err != nil {
				log.Println("Error closing db after failed ping", err)
				log.Println("Trying to connect again")
				time.Sleep(sleepTime)
				continue
			}
			log.Println("Unable to ping DB", err)
			log.Println("Waiting before attempting to reconnect")
			time.Sleep(sleepTime)
			continue
		}
		// keep looping till we get past all the error checks
		return db
	}
}

func createConnString(host string, user string, pass string, db string) string {
	return "host=" + host + " user=" + user + " password=" + pass + " dbname=" + db + " sslmode=disable"
}
