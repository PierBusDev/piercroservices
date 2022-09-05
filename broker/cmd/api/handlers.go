package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

func (c *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Reply from broker",
	}

	_ = c.writeJSON(w, http.StatusOK, payload)

}

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
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
	case "log":
		//c.logItem(w, requestPayload.Log)
		c.logEventOnRabbit(w, requestPayload.Log)
	case "mail":
		c.sendMail(w, requestPayload.Mail)
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

func (c *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		c.errorJSON(w, err)
		return
	}

	logServiceURL := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("[logItem]error while creating request to log service")
		c.errorJSON(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("[logItem]error while calling log service")
		c.errorJSON(w, err)
		return
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		c.errorJSON(w, errors.New("error in log service, status code is not StatusAccepted but "+response.Status))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Log entry created"

	c.writeJSON(w, http.StatusAccepted, payload)
}

func (c *Config) sendMail(w http.ResponseWriter, mail MailPayload) {
	jsonData, err := json.MarshalIndent(mail, "", "\t")
	if err != nil {
		c.errorJSON(w, err)
		return
	}

	//call mailservice
	mailServiceUrl := "http://mail-service/send"
	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		c.errorJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		c.errorJSON(w, errors.New("error in mail service, status code is not StatusAccepted but "+response.Status))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Mail sent successfully to: " + mail.To + " from " + mail.From
	c.writeJSON(w, http.StatusAccepted, payload)
}

func (c *Config) logEventOnRabbit(w http.ResponseWriter, l LogPayload) {
	err := c.pushToQueue(l.Name, l.Data)
	if err != nil {
		c.errorJSON(w, errors.New("error while pushing to RABBITMQ queue"))
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Logged entry via RABBITMQ"

	c.writeJSON(w, http.StatusAccepted, payload)
}

func (c *Config) pushToQueue(name, message string) error {
	emitter, err := event.NewEventEmitter(c.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	jsonPayload, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		return err
	}

	err = emitter.Push(string(jsonPayload), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}
