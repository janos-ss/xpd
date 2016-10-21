package xpd

import (
	ct "github.com/daviddengcn/go-colortext"
	"log"
	"fmt"
	"github.com/xpd-org/xpd/mail"
)

type ConsolePrinterListener struct{}

func (listener ConsolePrinterListener) OnDuplicates(post Post, oldPosts []Post) {
	ct.ResetColor()
	defer ct.ResetColor()
	ct.ChangeColor(ct.Red, true, ct.None, false)

	log.Printf("%s: possible cross-post on: %s (%s)", post.Feed.Id, post.Subject, post.Id)
	for _, oldPost := range oldPosts {
		log.Printf("  of: (%s:) %s (%s)", oldPost.Feed.Id, oldPost.Subject, oldPost.Id)
	}
	log.Println("------------------------------------------------------------")
}

type MailerListener struct {
	Mailer mail.Mailer
}

func (listener MailerListener) OnDuplicates(post Post, oldPosts []Post) {
	message := formatAsEmail(post, oldPosts)

	err := listener.Mailer.Send(message)
	if err != nil {
		log.Printf("smtp error: %s", err)
	}
}

func formatAsEmail(post Post, oldPosts []Post) string {
	message := fmt.Sprintf("%s: possible cross-post on: %s (%s)\n", post.Feed.Id, post.Subject, post.Id)
	for _, oldPost := range oldPosts {
		message += fmt.Sprintf("  of: (%s:) %s (%s)\n", oldPost.Feed.Id, oldPost.Subject, oldPost.Id)
	}
	return message
}
