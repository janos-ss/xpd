package xpd

import (
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"os"
	"io"
	"errors"
)

type RssReader struct {
	uri      string
	feed     *Feed
	rssFeed  *rss.Feed
	newPosts []Post
}

func NewRssReader(uri string, feed *Feed) *RssReader {
	reader := RssReader{uri, feed, nil, nil}
	timeout := 5
	reader.rssFeed = rss.New(timeout, true, reader.chanHandler, reader.itemHandler)
	return &reader
}

func (reader *RssReader) chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func (reader *RssReader) itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	fmt.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)

	posts := make([]Post, 0)
	for _, item := range (newitems) {
		post := Post{
			Url: item.Id,
			Author: item.Author.Name,
			Subject: item.Title,
			Body: item.Description,
		}
		posts = append(posts, post)
	}

	reader.newPosts = posts
}

func (reader *RssReader) GetNewPosts() []Post {
	reader.newPosts = nil

	if err := reader.rssFeed.Fetch(reader.uri, charsetReader); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %s: %s\n", reader.uri, err)
		return []Post{}
	}

	reader.feed.Posts = append(reader.feed.Posts, reader.newPosts...)

	return reader.newPosts
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" || charset == "iso-8859-1" {
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
}
