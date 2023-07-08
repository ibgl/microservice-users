package user

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	sender   string
	host     string
	port     string
	password string
}

func NewMailer(sender, host, port, pass string) *Mailer {
	return &Mailer{
		sender:   sender,
		host:     host,
		port:     port,
		password: pass,
	}
}

func (h *Mailer) Send(recipient, text string) error {
	return smtp.SendMail(fmt.Sprintf("%s:%s", h.host, h.port),
		smtp.PlainAuth("", h.sender, h.password, h.host),
		h.sender, []string{recipient}, []byte(text))
}
