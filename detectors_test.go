package xpd

import (
	"reflect"
	"testing"
)

func Test_SameBodyDetector_FindDuplicates_finds_same_body(t *testing.T) {
	body := "some text"
	differentBody := body + " blah"

	post := Post{Body: body}

	var repo PostRepository = NewPostRepository()
	repo.Add(post)
	repo.Add(Post{Body: differentBody})

	var detector Detector = SameBodyDetector{}
	if !reflect.DeepEqual(detector.FindDuplicates(post, []Post{post}), []Post{post}) {
		t.Fatal("same-body-detector should find only the match")
	}
}

func Test_splitToWords(t *testing.T) {
	s := "   @#$@hello THERE 4324%%%$# ouch  "
	if actual, expected := splitToWords(s), []string{"hello", "there", "ouch"}; !reflect.DeepEqual(actual, expected) {
		t.Fatalf("got %s; expected %s", actual, expected)
	}
}

func Test_wordCounts(t *testing.T) {
	s := "Hello World hello again"

	expected := &wordCountMap{
		counts: map[string]int{
			"hello": 2,
			"world": 1,
			"again": 1,
		},
		total: 4,
	}

	if actual := newWordCountMap(s); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("got %#v; expected %#v", actual, expected)
	}
}

func Test_similarEnoughCounts(t *testing.T) {
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
	if other := base + calcRatio(base, 1.1*limitRatio); similarCounts(base, other, limitRatio) {
		t.Errorf("got %d and %d are similar enough, but should _not_ be", base, other)
	}
	if other := base - calcRatio(base, 1.1*limitRatio); similarCounts(base, other, limitRatio) {
		t.Errorf("got %d and %d are similar enough, but should _not_ be", base, other)
	}
}

func Test_wordCountDiffs(t *testing.T) {
	first := &wordCountMap{
		counts: map[string]int{
			"hello": 7,
			"world": 13,
			"again": 17,
		},
		total: 7 + 13 + 17,
	}
	second := &wordCountMap{
		counts: map[string]int{
			"welcome": 23,
			"new":     29,
			"world":   31,
		},
		total: 23 + 29 + 31,
	}

	expectedDiffsLeft := float64(7+17) + float64(31-13)/2
	expectedDiffsRight := float64(23+29) + float64(31-13)/2

	if actual := calcWordCountDiffs(first, second); actual != expectedDiffsLeft {
		t.Errorf("got %f; expected %f", actual, expectedDiffsLeft)
	}

	if actual := calcWordCountDiffs(second, first); actual != expectedDiffsRight {
		t.Errorf("got %f; expected %f", actual, expectedDiffsRight)
	}
}

func Test_SimilarWordCountDetector_with_rearranged_words(t *testing.T) {
	post := Post{Body: "The quick brown fox jumps over the lazy dog"}
	rearranged := []Post{{Body: "the lazy dog The quick brown fox jumps over"}}

	if !reflect.DeepEqual(SimilarWordCountDetector{}.FindDuplicates(post, rearranged), rearranged) {
		t.Errorf("got '%v' not a duplicate of '%v', but it should be", rearranged[0].Body, post.Body)
	}
}

func Test_SimilarWordCountDetector_with_deleted_words(t *testing.T) {
	post := Post{Body: "The quick brown fox jumps over the lazy dog filler filler"}
	deleted := []Post{{Body: "The quick brown fox over the lazy dog filler filler"}}

	if !reflect.DeepEqual(SimilarWordCountDetector{}.FindDuplicates(post, deleted), deleted) {
		t.Errorf("got '%v' not a duplicate of '%v', but it should be", deleted[0].Body, post.Body)
	}
}

func Test_SimilarWordCountDetector_with_added_words(t *testing.T) {
	post := Post{Body: "The quick brown fox jumps over the lazy dog filler filler"}
	added := []Post{{Body: "The quick brown fox jumps over the dumb lazy dog filler filler"}}

	if !reflect.DeepEqual(SimilarWordCountDetector{}.FindDuplicates(post, added), added) {
		t.Errorf("got '%v' not a duplicate of '%v', but it should be", added[0].Body, post.Body)
	}
}
