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
	if !similarEnoughCounts(113, 115) {
		t.Error("got 113, 115 are not similar enough, but should be")
	}
	if !similarEnoughCounts(1130, 1200) {
		t.Error("got 1130, 1200 are not similar enough, but should be")
	}
	if similarEnoughCounts(1130, 1500) {
		t.Error("got 1130, 1500 are similar enough, but should not be")
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
