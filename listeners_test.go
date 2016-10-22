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

func Test_summaryOfPost(t*testing.T) {
	post := Post{Id: "feed1-1", Subject: "sub1-1", Feed: &Feed{Id: "feed1"}}
	actual := summaryOfPost(post)
	expected := "feed=feed1; subject=sub1-1; id=feed1-1"

	if actual != expected {
		t.Fatalf("got: %s\nexpected: %s", actual, expected)
	}
}
