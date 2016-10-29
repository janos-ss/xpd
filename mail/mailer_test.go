package mail

import (
	"testing"
)

func Test_MockMailer_Send(t *testing.T) {
	message := "hello world"

	mailer := &MockMailer{}
	mailer.Send(message)

	if mailer.Message != message {
		t.Fatalf("got %s; expected %s", mailer.Message, message)
	}
}

func Test_NullMailer_send(t *testing.T) {
	mailer := NullMailer{}
	if mailer.Send("some message") == nil {
		t.Fatal("got success for NullMailer.Send; expected to always fail")
	}
}

func Test_GmailMailer_String_should_hide_password(t *testing.T) {
	pass := "pass"
	mailer := GmailMailer{From: "from", Pass: pass}

	expected := `mail.GmailMailer{From:"*", Pass:"*", Recipient:"*", Subject:""}`

	if s := mailer.String(); s != expected {
		t.Fatalf("got %s, expected %s", s, expected)
	}
	if mailer.Pass != pass {
		t.Fatalf("got mailer.Pass=%s; expected %s", mailer.Pass, pass)
	}
}
