package main

import (
	"net/http"
)

func (c *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Reply from broker",
	}

	_ = c.writeJSON(w, http.StatusOK, payload)

}
