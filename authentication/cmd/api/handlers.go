package main

import (
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
	
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Successfully authenticated user: %s", user.Email),
		Data:    user,
	}

	c.writeJSON(w, http.StatusAccepted, payload)
}
