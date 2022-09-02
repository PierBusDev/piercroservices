package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

func (c *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Reply from broker",
	}

	_ = c.writeJSON(w, http.StatusOK, payload)

}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
}

func (c *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := c.readJSON(w, r, &requestPayload)
	if err != nil {
		c.errorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		c.authenticate(w, requestPayload.Auth)
	default:
		c.errorJSON(w, errors.New("unkown action"))
	}
}

func (c *Config) authenticate(w http.ResponseWriter, payload AuthPayload) {
	jsonData, _ := json.MarshalIndent(payload, "", "\t")

	//calling authservice
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		c.errorJSON(w, err)
		return
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	//check expected status code
	if response.StatusCode == http.StatusUnauthorized {
		c.errorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		c.errorJSON(w, errors.New("error in authentication service"))
		return
	}

	var res jsonResponse
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		c.errorJSON(w, err)
		return
	}

	if res.Error { //remember it's a bool
		c.errorJSON(w, errors.New(res.Message), http.StatusUnauthorized)
		return
	}

	//if we are HERE we have a valid login
	var retPayload jsonResponse
	retPayload.Error = false
	retPayload.Message = "Authenticated, Login successful"
	retPayload.Data = res.Data

	c.writeJSON(w, http.StatusAccepted, retPayload)
}
