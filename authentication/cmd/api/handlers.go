package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (c *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := c.readJSON(w, r, &requestPayload)
	if err != nil {
		c.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := c.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		c.errorJSON(w, errors.New("Invalid Credentials"), http.StatusBadRequest)
		return
	}

	validPsw, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !validPsw {
		c.errorJSON(w, errors.New("Invalid Credentials"), http.StatusBadRequest)
		return
	}

	//logging authentication
	err = c.logRequest("authentication", fmt.Sprintf("User %s authenticated", user.Email))
	if err != nil {
		c.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	//creating and sending back response
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Successfully authenticated user: %s", user.Email),
		Data:    user,
	}

	c.writeJSON(w, http.StatusAccepted, payload)
}

func (c *Config) logRequest(name, data string) error {
	entry := struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}{
		Name: name,
		Data: data,
	}

	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		return err
	}

	logServiceUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}
