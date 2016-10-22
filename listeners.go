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

	log.Printf("possible cross-post(s):\n%s", summaryOfDups(post, oldPosts))
}

type MailerListener struct {
	Mailer mail.Mailer
}

func (listener MailerListener) OnDuplicates(post Post, oldPosts []Post) {
	message := summaryOfDups(post, oldPosts)

	err := listener.Mailer.Send(message)
	if err != nil {
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
