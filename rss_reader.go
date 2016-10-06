package xpd

import (
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"os"
	"io"
	"errors"
)

type RssReader struct {
	Uri      string
	Feed     Feed
	RssFeed  *rss.Feed
	NewPosts []Post
}

func (reader *RssReader) Init() {
	timeout := 5
	reader.RssFeed = rss.New(timeout, true, reader.chanHandler, reader.itemHandler)
}

func (reader *RssReader) chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func (reader *RssReader) itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)

	posts := make([]Post, len(newitems))
	for _, item := range (newitems) {
		post := Post{
			Url: item.Id,
			Author: item.Author.Name,
			Subject: item.Title,
			Body: item.Description,
		}
		posts = append(posts, post)
	}

	reader.NewPosts = posts
}

func (reader *RssReader) GetNewPosts() []Post {
	reader.NewPosts = nil

	if err := reader.RssFeed.Fetch(reader.Uri, charsetReader); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %s: %s\n", reader.Uri, err)
		return []Post{}
	}

	return reader.NewPosts
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" || charset == "iso-8859-1" {
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
}
