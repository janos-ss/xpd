package xpd

import "fmt"

type consolePrinterListener struct{}

func (listener consolePrinterListener) onDuplicates(post Post, oldPosts []Post) {
	fmt.Printf("possible cross-post: %s (%s)\n", post.Subject, post.Id)
	for _, oldPost := range(oldPosts) {
		fmt.Printf("  of: %s (%s)\n", oldPost.Subject, oldPost.Id)
	}
	fmt.Println()
}
