package main

import (
	"authentication/data"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const port = "80"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting auth service")
	//connect to DB
	conn := connectToDb()
	if conn == nil {
		log.Fatal("can't connect to postgres...")
	}
	//set up config
	config := Config{
		DB:     conn,
		Models: data.New(conn),
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: config.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

var counts int64

const maxTries = 10

func connectToDb() *sql.DB {
	dsn := os.Getenv("DSN")
	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("postgres not yet ready...")
			counts++
		} else {
			log.Println("connected successfully to postgres")
			return connection
		}

		if counts > maxTries {
			log.Println("Failed to connect after", maxTries, "tries")
			return nil
		}
		log.Println("Gonna retry in two seconds")
		time.Sleep(time.Second * 2)
	}
}
