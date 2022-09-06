package main

import (
	"context"
	"log"
	"logger/data"
	"time"
)

type RPCserver struct{}

type RPCPayload struct {
	Name string
	Data string
}

//LogInfo writes the payload to mongo
func (r *RPCserver) LogInfo(payload RPCPayload, response *string) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.Background(), data.LogEntry{
		Name:      payload.Name,
		Data:      payload.Data,
		CreatedAt: time.Now(),
	})
	if err != nil {
		log.Println("RPC->LogInfo, failed to insert log entry, ", err)
		return err
	}

	*response = "Payload processed via RPC: " + payload.Name
	return nil
}
