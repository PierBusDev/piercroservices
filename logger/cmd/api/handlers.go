package main

import (
	"log"
	"logger/data"
	"net/http"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (c *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	//read json
	var requestPayload JSONPayload
	err := c.readJSON(w, r, &requestPayload)
	if err != nil {
		log.Println("[writelog]something went wrong while reading the json payload")
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}
	err = c.Models.LogEntry.Insert(event)
	if err != nil {
		log.Println("[writelog]something went wrong while inserting the log entry in the db")
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: "Successfully inserted log entry",
		Data:    event,
	}

	c.writeJSON(w, http.StatusAccepted, response)
}
