package xpd

import (
	"fmt"
	"github.com/xpd-org/xpd/mail"
	"log"
)

type ConsolePrinterListener struct{}

func (listener ConsolePrinterListener) OnCrossPost(post Post, posts []Post) {
	log.Printf("possible cross-post:\n%s", summaryOfPosts(post, posts))
}

func (listener ConsolePrinterListener) OnDuplicate(post Post, posts []Post) {
	log.Printf("possible duplicate:\n%s", summaryOfPosts(post, posts))
}

type MailerListener struct {
	Mailer mail.Mailer
}

func (listener MailerListener) OnCrossPost(post Post, posts []Post) {
	log.Println("sending email about cross-post:", post.Id)
	listener.send("possible cross-post:", post, posts)
}

func (listener MailerListener) OnDuplicate(post Post, posts []Post) {
	log.Println("sending email about duplicate post:", post.Id)
	listener.send("possible duplicate:", post, posts)
}

func (listener MailerListener) send(subject string, post Post, posts []Post) {
	// start message with empty line to avoid interpretation as header fields
	message := "\n\n"
	message += subject + "\n\n"
	message += summaryOfPosts(post, posts)

	if err := listener.Mailer.Send(message); err != nil {
		log.Printf("smtp error: %s", err)
	}
}

func summaryOfPosts(post Post, oldPosts []Post) string {
	summary := fmt.Sprintf("%s\n", summaryOfPost(post))
	for _, old := range oldPosts {
		summary += fmt.Sprintf("  of: %s\n", summaryOfPost(old))
	}
	return summary
}

func summaryOfPost(post Post) string {
	return fmt.Sprintf("feed=%s; id=%s; subject=%s", post.Feed.Id, post.Id, post.Subject)
}
