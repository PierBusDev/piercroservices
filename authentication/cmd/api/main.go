package main

import (
	"authentication/data"
	"database/sql"
	"log"
	"net/http"
)

const port = "80"

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting auth service")
	//connect to DB

	//set up config
	config := Config{}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: config.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
