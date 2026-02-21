package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func SetupDatabase() {
	dsn := "user=pqgo dbname=pqgo sslmode=verify-full"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Connected to database")
	db.Close()
}
