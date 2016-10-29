package xpd

import (
	rss "github.com/jteeuwen/go-pkg-rss"
	"testing"
)

func newDummyItem() *rss.Item {
	link1 := rss.Link{}
	link2 := rss.Link{}
	return &rss.Item{Links: []*rss.Link{&link1, &link2}}
}

func Test_extractPostId_should_return_id_if_exists(t *testing.T) {
	id := "dummyId"
	item := newDummyItem()
	item.Id = id
	if actual := extractPostId(item); actual != id {
		t.Fatalf("got %#v, expected %#v", actual, id)
	}
}

func Test_extractPostId_should_return_first_link_if_exists_and_id_missing(t *testing.T) {
	id := "dummyId"
	item := newDummyItem()
	item.Links[0].Href = id
	if actual := extractPostId(item); actual != id {
		t.Fatalf("got %#v, expected %#v", actual, id)
	}
}

func Test_extractPostId_should_return_empty_string_when_no_id_no_link(t *testing.T) {
	id := ""
	item := newDummyItem()
	if actual := extractPostId(item); actual != id {
		t.Fatalf("got %#v, expected %#v", actual, id)
	}
}

func Test_itemHandler(t *testing.T) {
	reader := NewRssReader("dummy url", Feed{}).(*rssReader)

	reader.itemHandler(nil, nil, []*rss.Item{{}})
	if len(reader.newPosts) != 1 {
		t.Fatalf("got %d posts, expected 1", len(reader.newPosts))
	}

	reader.itemHandler(nil, nil, []*rss.Item{{}, {}})
	if len(reader.newPosts) != 2 {
		t.Fatalf("got %d posts, expected 2", len(reader.newPosts))
	}
}
