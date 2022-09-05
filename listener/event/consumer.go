package event

import (
	"bytes"
	"encoding/json"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		log.Println("something went wrong in NewConsumer, ", err)
		return Consumer{}, err
	}

	return consumer, nil
}

func (cons *Consumer) setup() error {
	channel, err := cons.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (cons *Consumer) Listen(topics []string) error {
	channel, err := cons.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	q, err := declareRandomQueue(channel)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		err = channel.QueueBind(
			q.Name,       // queue name
			topic,        // routing key
			"logs_topic", // exchange
			false,
			nil)
		if err != nil {
			log.Println("failed to bind queue to exchange, ", err)
			return err
		}
	}

	//consume
	messages, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("failed to consume messages, ", err)
		return err
	}

	forever := make(chan bool)
	go func() {
		for mess := range messages {
			log.Printf("received message: %s", mess.Body)
			var payload Payload
			err := json.Unmarshal(mess.Body, &payload)
			if err != nil {
				log.Println("failed to unmarshal message, ", err)
				return
			}

			go handlePayload(payload)
		}
	}()

	log.Printf("Waiting for message on [exchange]: %s, [queue]: %s", "logs_topic", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	log.Println("handling payload: ", payload)
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Println("failed to log event, ", err)
		}
	case "auth":
		//TODO
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println("failed to log event, ", err)
		}
	}
}

func logEvent(entry Payload) error {
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		return err
	}

	logServiceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("[logItem]error while creating request to log service")
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("[logItem]error while calling log service")
		return err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}
