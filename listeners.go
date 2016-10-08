package xpd

import (
	"log"
	ct "github.com/daviddengcn/go-colortext"
)

type consolePrinterListener struct{}

func (listener consolePrinterListener) onDuplicates(post Post, oldPosts []Post) {
	ct.ResetColor()
	defer ct.ResetColor()
	ct.ChangeColor(ct.Red, true, ct.None, false)
	log.Printf("possible cross-post on %s : %s (%s)\n", post.Feed.Id, post.Subject, post.Id)
	for _, oldPost := range oldPosts {
		log.Printf("  of: %s %s (%s)\n", oldPost.Feed.Id, oldPost.Subject, oldPost.Id)
	}
	log.Println("----")
}
