package main

import (
	"fmt"
	"github.com/xpd-org/xpd"
)

func main() {
	uri := "http://stackoverflow.com/feeds/tag?tagnames=sonarqube&sort=newest"
	feed := xpd.Feed{}
	rssReader := xpd.NewRssReader(uri, &feed)
	getAndPrintNewPosts(rssReader)
}

func getAndPrintNewPosts(reader xpd.FeedReader) {
	for _, post := range reader.GetNewPosts() {
		fmt.Println(post.Author)
	}
}
