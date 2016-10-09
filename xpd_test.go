package xpd

import (
	"testing"
	"reflect"
)

func Test_adding_to_repo(t*testing.T) {
	var repo PostRepository = newSimplePostRepository()
	repo.add(Post{})

	if len(repo.findRecent()) == 0 {
		t.Errorf("PostRepository should not be empty after post added")
	}
}

func Test_sameBodyDetector_findDuplicates_finds_same_body(t*testing.T) {
	body := "some text"
	differentBody := body + " blah"

	post := Post{Body: body}

	var repo PostRepository = newSimplePostRepository()
	repo.add(post)
	repo.add(Post{Body: differentBody})

	var detector Detector = sameBodyDetector{}
	if ! reflect.DeepEqual(detector.findDuplicates(post, []Post{post}), []Post{post}) {
		t.Errorf("same-body-detector should find only the match")
	}
}

func Test_ellipsize_someString_15_is_someString(t*testing.T) {
	s := "someString"
	if actual := ellipsize(s, 15); actual != s {
		t.Fatalf("got %s; expected %s", actual, s)
	}
}

func Test_ellipsize_someString_7_is_somedots(t*testing.T) {
	s := "someString"
	if actual, expected := ellipsize(s, 7), "some..."; actual != expected {
		t.Fatalf("got %s; expected %s", actual, expected)
	}
}

func Test_splitToWords(t*testing.T) {
	s := "   @#$@hello THERE 4324%%%$# ouch  "
	if actual, expected := splitToWords(s), []string{"hello", "there", "ouch"}; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("got %s; expected %s", actual, expected)
	}
}

func Test_wordCounts(t*testing.T) {
	s := "Hello World hello again"

	expected := wordCountMap{
		"hello": 2,
		"world": 1,
		"again": 1,
	}
	expectedTotal := 4

	if actual, actualTotal := calcWordCounts(s); !reflect.DeepEqual(actual, expected) || actualTotal != expectedTotal {
		t.Fatalf("got %v, %d; expected %v, %d", actual, actualTotal, expected, expectedTotal)
	}
}

func Test_similarEnoughCounts(t*testing.T) {
	limitRatio := 0.1
	base := 123
	if other := base; !similarCounts(base, other, limitRatio) {
		t.Errorf("got %d and %d are _not_ similar enough, but should be", base, other)
	}
	if other := base + calcRatio(base, limitRatio); !similarCounts(base, other, limitRatio) {
		t.Errorf("got %d and %d are _not_ similar enough, but should be", base, other)
	}
	if other := base - calcRatio(base, limitRatio); !similarCounts(base, other, limitRatio) {
		t.Errorf("got %d and %d are _not_ similar enough, but should be", base, other)
	}
	if other := base + calcRatio(base, 1.1 * limitRatio); similarCounts(base, other, limitRatio) {
		t.Errorf("got %d and %d are similar enough, but should _not_ be", base, other)
	}
	if other := base - calcRatio(base, 1.1 * limitRatio); similarCounts(base, other, limitRatio) {
		t.Errorf("got %d and %d are similar enough, but should _not_ be", base, other)
	}
}

func Test_wordCountDiffs(t*testing.T) {
	first := wordCountMap{
		"hello": 7,
		"world": 13,
		"again": 17,
	}
	second := wordCountMap{
		"welcome": 23,
		"new": 29,
		"world": 31,
	}

	expectedDiffsLeft := float64(7 + 17) + float64(31 - 13) / 2
	expectedDiffsRight := float64(23 + 29) + float64(31 - 13) / 2

	if actual := calcWordCountDiffs(first, second); actual != expectedDiffsLeft {
		t.Errorf("got %f; expected %f", actual, expectedDiffsLeft)
	}

	if actual := calcWordCountDiffs(second, first); actual != expectedDiffsRight {
		t.Errorf("got %f; expected %f", actual, expectedDiffsRight)
	}
}

func Test_similarWordCountDetector_with_rearranged_words(t*testing.T) {
	post := Post{Body: "The quick brown fox jumps over the lazy dog"}
	rearranged := []Post{Post{Body: "the lazy dog The quick brown fox jumps over"}}

	if !reflect.DeepEqual(similarWordCountDetector{}.findDuplicates(post, rearranged), rearranged) {
		t.Errorf("got '%v' not a duplicate of '%v', but it should be", rearranged[0].Body, post.Body)
	}
}

func Test_similarWordCountDetector_with_deleted_words(t*testing.T) {
	post := Post{Body: "The quick brown fox jumps over the lazy dog filler filler"}
	deleted := []Post{Post{Body: "The quick brown fox over the lazy dog filler filler"}}

	if !reflect.DeepEqual(similarWordCountDetector{}.findDuplicates(post, deleted), deleted) {
		t.Errorf("got '%v' not a duplicate of '%v', but it should be", deleted[0].Body, post.Body)
	}
}

func Test_similarWordCountDetector_with_added_words(t*testing.T) {
	post := Post{Body: "The quick brown fox jumps over the lazy dog filler filler"}
	added := []Post{Post{Body: "The quick brown fox jumps over the dumb lazy dog filler filler"}}

	if !reflect.DeepEqual(similarWordCountDetector{}.findDuplicates(post, added), added) {
		t.Errorf("got '%v' not a duplicate of '%v', but it should be", added[0].Body, post.Body)
	}
}

func Test_simpleDetectorRegistry(t*testing.T) {
	detector := similarWordCountDetector{}

	reg := newSimpleDetectorRegistry()
	reg.register(detector)

	if d := reg.get("similarWordCountDetector"); d != detector {
		t.Errorf("got %#v, expected %#v", d, detector)
	}

	assertPanic(t, "did not crash on unknown Detector, but it should have", func() {
		reg.get("nonexistent")
	})
}

func Test_getDetectors(t*testing.T) {
	reg := newSimpleDetectorRegistry()
	reg.register(sameBodyDetector{})
	reg.register(similarWordCountDetector{})

	detectors := getDetectors(reg, []string{"sameBodyDetector", "similarWordCountDetector"})
	expected := []Detector{sameBodyDetector{}, similarWordCountDetector{}}

	if !reflect.DeepEqual(detectors, expected) {
		t.Errorf("got %#v, expected %#v", detectors, expected)
	}

	assertPanic(t, "did not crash on unknown Detector, but it should have", func() {
		getDetectors(reg, []string{"nonexistent"})
	})
}

func assertPanic(t *testing.T, message string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(message)
		}
	}()
	f()
}

func Test_createContext(t*testing.T) {
	config := Config{
		Feeds: []Feed{
			Feed{Id: "dummy1", Url: "dummy1"},
			Feed{Id: "dummy2", Url: "dummy2"},
		},
		DetectorNames: []string{"sameBodyDetector"},
	}
	context := createContext(config)

	if len(context.readers) != len(config.Feeds) {
		t.Errorf("got different number of feed readers than specified feeds; %#v <- %#v", context.readers, config.Feeds)
	}
	if len(context.readers) < 1 {
		t.Error("got no feed readers, expected at least 1")
	}

	if len(context.detectors) != len(config.DetectorNames) {
		t.Errorf("got different number of detectors than specified names; %#v <- %#v", context.detectors, config.DetectorNames)
	}
	if len(context.detectors) < 1 {
		t.Error("got no detectors, expected at least 1")
	}

	if len(context.listeners) < 1 {
		t.Error("got no listeners, expected at least 1")
	}
	if context.postRepository == nil {
		t.Error("got nil PostRepository, expected non-nil")
	}
}
