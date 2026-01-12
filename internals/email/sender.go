package email

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type Sender struct {
	host     string
	port     int
	username string
	password string
}


func NewSender(host string, port int, user, pass string) *Sender {
	return &Sender{
		host:     host,
		port:     port,
		username: user,
		password: pass,
	}
}

func (s *Sender) Send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)
d.TLSConfig = &tls.Config{
		ServerName: s.host,
	}

	return d.DialAndSend(m)
}

