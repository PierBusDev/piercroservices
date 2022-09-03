package main

import (
	"bytes"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"time"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string //email address
	FromName    string
	To          string
	Subject     string
	Attachments []string //pathnames
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	//setting defaults in case msg has those fields empty
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	msg.DataMap = map[string]any{ //will be used when rendering the template
		"message": msg.Data,
	}

	//creating both an html and a plaintext version of the mail we want to send
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}
	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	//starting the server
	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	//creating the email
	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	//adding attachments
	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			email.AddAttachment(attachment)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}
	return nil
}

//buildHTMLMessage builds an email message in html format
func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	mailTemplate := "./templates/mail.html.gohtml"
	t, err := template.New("email-html").ParseFiles(mailTemplate)
	if err != nil {
		return "", err
	}

	var tmpl bytes.Buffer
	if err = t.ExecuteTemplate(&tmpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tmpl.String()
	//adding css
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

//buildPlainTextMessage builds an email message in simple plain text format
func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	mailTemplate := "./templates/mail.plaintext.gohtml"
	t, err := template.New("email-plain").ParseFiles(mailTemplate)
	if err != nil {
		return "", err
	}

	var tmpl bytes.Buffer
	if err = t.ExecuteTemplate(&tmpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	plaintextMessage := tmpl.String()

	return plaintextMessage, nil
}

func (m *Mail) inlineCSS(stringedMsg string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(stringedMsg, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncryption(s string) mail.Encryption {
	switch s {
	case "ssl":
		return mail.EncryptionSSLTLS
	case "tls":
		return mail.EncryptionSTARTTLS
	default:
		return mail.EncryptionNone
	}
}
