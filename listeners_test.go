package xpd

import (
	"testing"
)

func Test_consolePrinterListener_should_crash_on_Post_without_Feed(t *testing.T) {
	postWithoutFeed := Post{Subject: "dummyPost"}
	assertPanic(t, "did not crash on Post without Feed, but it should have", func() {
		ConsolePrinterListener{}.OnDuplicates(postWithoutFeed, []Post{{}})
	})
}

func Test_consolePrinterListener_should_crash_on_duplicate_Post_without_Feed(t *testing.T) {
	postWithFeed := Post{Subject: "dummyPost", Feed: &Feed{Id: "dummyFeed"}}
	postWithoutFeed := Post{}

	assertPanic(t, "did not crash on Post without Feed, but it should have", func() {
		ConsolePrinterListener{}.OnDuplicates(postWithFeed, []Post{postWithoutFeed})
	})
}

func Test_consolePrinterListener_happy_path(t *testing.T) {
	postWithFeed := Post{Subject: "dummyPost", Feed: &Feed{Id: "dummyFeed"}}
	ConsolePrinterListener{}.OnDuplicates(postWithFeed, []Post{postWithFeed})
}

func Test_formatAsEmail(t*testing.T) {
	post := Post{Id: "feed1-1", Subject: "feed1-1", Feed: &Feed{Id: "feed1"}}
	old := []Post{
		{Id: "feed2-1", Subject: "feed2-1", Feed: &Feed{Id: "feed2"}},
		{Id: "feed3-1", Subject: "feed3-1", Feed: &Feed{Id: "feed3"}},
	}

	actual := formatAsEmail(post, old)
	expected := "feed1: possible cross-post on: feed1-1 (feed1-1)\n"
	expected += "  of: (feed2:) feed2-1 (feed2-1)\n"
	expected += "  of: (feed3:) feed3-1 (feed3-1)\n"

	if actual != expected {
		t.Fatalf("got: %s\nexpected: %s", actual, expected)
	}
}
