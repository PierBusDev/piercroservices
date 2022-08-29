package main

import (
	"log"
	"net/http"
)

const webPort = "9999"

type Config struct{}

func main() {
	app := Config{}

	log.Printf("starting broker service on port %s\n", webPort)
	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
