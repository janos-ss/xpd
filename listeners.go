package xpd

import (
	"log"
)

type consolePrinterListener struct{}

func (listener consolePrinterListener) onDuplicates(post Post, oldPosts []Post) {
	log.Printf("possible cross-post on %s : %s (%s)\n", post.Feed.Id, post.Subject, post.Id)
	for _, oldPost := range (oldPosts) {
		log.Printf("  of: %s %s (%s)\n", oldPost.Feed.Id, oldPost.Subject, oldPost.Id)
	}
	log.Println("----")
}
