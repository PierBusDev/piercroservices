package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"time"
)

const webPort = "9999"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	//connect to rabbitMQ
	rabbitConn, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()
	log.Println("successfully connected to rabbitMQ")

	app := Config{
		Rabbit: rabbitConn,
	}

	log.Printf("starting broker service on port %s\n", webPort)
	srv := &http.Server{
		Addr:    ":" + webPort,
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

var maxRetries = 10

func connect() (*amqp.Connection, error) {
	var counts int64
	backoff := 2 * time.Second
	var connection *amqp.Connection
	//connect to rabbitMQ (with various tries)
	for {
		c, err := amqp.Dial("amqp://admin:password@rabbitmq")
		if err != nil {
			fmt.Println("failed to connect to rabbitMQ at the moment... (probably not yet ready)")
			counts++
		} else {
			connection = c
			break
		}
		if counts > int64(maxRetries) {
			fmt.Printf("failed to connect to rabbitMQ after %s tries\n", maxRetries)
			return nil, err
		}
		backoff = backoff * 2
		log.Println("retrying in ", backoff)
		time.Sleep(backoff)
	}

	//return the connection
	return connection, nil
}
