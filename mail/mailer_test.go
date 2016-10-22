package mail

import "testing"

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
