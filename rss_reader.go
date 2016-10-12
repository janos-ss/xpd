package xpd

import (
	"errors"
	"fmt"
	ct "github.com/daviddengcn/go-colortext"
	rss "github.com/jteeuwen/go-pkg-rss"
	"io"
	"log"
	"os"
)

type rssReader struct {
	uri      string
	feed     Feed
	rssFeed  *rss.Feed
	newPosts []Post
}

func NewRssReader(uri string, feed Feed) FeedReader {
	reader := rssReader{uri: uri, feed: feed}
	timeout := 0
	reader.rssFeed = rss.New(timeout, true, reader.chanHandler, reader.itemHandler)
	return &reader
}

func (reader *rssReader) chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	ct.ResetColor()
	defer ct.ResetColor()
	ct.ChangeColor(ct.Blue, true, ct.None, false)
	log.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func (reader *rssReader) itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	ct.ResetColor()
	defer ct.ResetColor()
	ct.ChangeColor(ct.Green, true, ct.None, false)
	log.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)

	posts := make([]Post, len(newitems))
	for i, item := range newitems {
		post := Post{
			Id:      item.Id,
			Url:     item.Id,
			Author:  item.Author.Name,
			Subject: item.Title,
			Body:    item.Description,
			Feed:    &reader.feed,
		}
		posts[i] = post
	}

	reader.newPosts = posts
}

func (reader *rssReader) GetFeed() Feed {
	return reader.feed
}

func (reader *rssReader) FetchNewPosts() []Post {
	reader.newPosts = nil

	if err := reader.rssFeed.Fetch(reader.uri, charsetReader); err != nil {
		fmt.Fprintf(os.Stderr, "[e] %s: %s\n", reader.uri, err)
		return []Post{}
	}

	return reader.newPosts
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" || charset == "iso-8859-1" {
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
}
