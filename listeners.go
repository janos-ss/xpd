package xpd

import (
	"log"
)

type consolePrinterListener struct{}

func (listener consolePrinterListener) onDuplicates(post Post, oldPosts []Post) {
	log.Printf("possible cross-post: %s (%s)\n", post.Subject, post.Id)
	for _, oldPost := range (oldPosts) {
		log.Printf("  of: %s (%s)\n", oldPost.Subject, oldPost.Id)
	}
	log.Println("----")
}
