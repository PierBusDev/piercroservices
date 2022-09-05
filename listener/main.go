package main

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"listener/event"
	"log"
	"time"
)

func main() {
	//connect to rabbitMQ
	rabbitConn, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()
	log.Println("successfully connected to rabbitMQ")

	//start listening for messages and create a consumer
	log.Println("starting to listen for RABBITMQ messages")
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		log.Fatal(err)
	}

	//watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println("Something happened while trying to listen to topics ", err)
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
