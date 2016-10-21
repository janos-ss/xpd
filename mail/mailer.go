package mail

import (
	"github.com/SlyMarbo/gmail"
)

type Mailer interface {
	Send(message string) error
}

type MockMailer struct {
	message string
}

func (mailer *MockMailer) Send(message string) error {
	mailer.message = message
	return nil
}

type GmailMailer struct {
	From      string
	Pass      string
	Recipient string
	Subject   string
}

func (mailer GmailMailer) Send(message string) error {
	email := gmail.Compose(mailer.Subject, message)
	email.From = mailer.From
	email.Password = mailer.Pass

	// Defaults to "text/plain; charset=utf-8" if unset.
	//email.ContentType = "text/html; charset=utf-8"

	// Normally you'll only need one of these, but I thought I'd show both.
	email.AddRecipient(mailer.Recipient)

	return email.Send()
}
