package xpd

import (
	"fmt"
	"github.com/xpd-org/xpd/mail"
	"log"
)

type ConsolePrinterListener struct{}

func (listener ConsolePrinterListener) OnDuplicates(post Post, oldPosts []Post) {
	log.Printf("possible cross-post(s):\n%s", summaryOfDups(post, oldPosts))
}

type MailerListener struct {
	Mailer mail.Mailer
}

func (listener MailerListener) OnDuplicates(post Post, oldPosts []Post) {
	log.Println("sending email about post:", post.Id)

	// start message with empty line to avoid interpretation as header fields
	message := "\n\n"
	message += summaryOfDups(post, oldPosts)

	if err := listener.Mailer.Send(message); err != nil {
		log.Printf("smtp error: %s", err)
	}
}

func summaryOfDups(post Post, oldPosts []Post) string {
	summary := fmt.Sprintf("%s\n", summaryOfPost(post))
	for _, old := range oldPosts {
		summary += fmt.Sprintf("  of: %s\n", summaryOfPost(old))
	}
	return summary
}

func summaryOfPost(post Post) string {
	return fmt.Sprintf("feed=%s; subject=%s; id=%s", post.Feed.Id, post.Subject, post.Id)
}
