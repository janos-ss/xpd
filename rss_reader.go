package xpd

import (
	"errors"
	rss "github.com/jteeuwen/go-pkg-rss"
	"io"
	"log"
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
	log.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func (reader *rssReader) itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	log.Printf("%d new item(s) in %s\n", len(newitems), feed.Url)

	posts := make([]Post, len(newitems))
	for i, item := range newitems {
		id := extractPostId(item)
		post := Post{
			Id:      id,
			Url:     id,
			Author:  item.Author.Name,
			Subject: item.Title,
			Body:    item.Description,
			Feed:    &reader.feed,
		}
		posts[i] = post
	}

	reader.newPosts = posts
}

// RSS feeds store the post id in different non-standard places
func extractPostId(item *rss.Item) string {
	if len(item.Id) > 0 {
		return item.Id
	}
	for _, link := range item.Links {
		if len(link.Href) > 0 {
			return link.Href
		}
	}
	return ""
}

func (reader *rssReader) GetFeed() Feed {
	return reader.feed
}

func (reader *rssReader) FetchNewPosts() []Post {
	reader.newPosts = nil

	if err := reader.rssFeed.Fetch(reader.uri, charsetReader); err != nil {
		log.Printf("error: %s: %s\n", reader.uri, err)
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
