package main

import (
	"fmt"
	"github.com/janos-ss/xpd"
)

func main() {
	uri := "http://stackoverflow.com/feeds/tag?tagnames=sonarqube&sort=newest"
	rssReader := xpd.NewRssReader(uri, xpd.Feed{})
	getAndPrintNewPosts(rssReader)
}

func getAndPrintNewPosts(reader xpd.FeedReader) {
	for _, post := range reader.GetNewPosts() {
		fmt.Println(post.Author)
	}
}
