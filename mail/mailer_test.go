package mail

import "testing"

func Test_MockMailer_Send(t *testing.T) {
	message := "hello world"

	mailer := &MockMailer{}
	mailer.Send(message)

	if mailer.message != message {
		t.Fatalf("got %s; expected %s", mailer.message, message)
	}
}
