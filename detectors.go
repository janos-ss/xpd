package xpd

import (
	"strings"
	"regexp"
)

type sameBodyDetector struct{}

func (detector sameBodyDetector) findDuplicates(post Post, oldPosts []Post) []Post {
	duplicates := make([]Post, 0)
	for _, oldPost := range (oldPosts) {
		if post.Body == oldPost.Body {
			duplicates = append(duplicates, oldPost);
		}
	}
	return duplicates
}

type similarWordCountDetector struct{}

type wordCountMap map[string]int

func (detector similarWordCountDetector) findDuplicates(post Post, oldPosts []Post) []Post {
	wordCounts, total := calcWordCounts(post.Body)

	duplicates := make([]Post, 0)
	for _, oldPost := range (oldPosts) {
		otherWordCounts, otherTotal := calcWordCounts(oldPost.Body)
		if similarEnough(wordCounts, total, otherWordCounts, otherTotal) {
			duplicates = append(duplicates, oldPost);
		}
	}
	return duplicates
}

func calcWordCounts(text string) (wordCountMap, int) {
	wordCounts := make(wordCountMap)
	// TODO
	return wordCounts, 0
}

func splitToWords(text string) []string {
	r := regexp.MustCompile("[^a-z]+")
	abc :=strings.TrimSpace(r.ReplaceAllString(strings.ToLower(text), " "))
	return strings.Split(abc, " ")
}

func similarEnough(wordCounts wordCountMap, total int, otherWordCounts wordCountMap, otherTotal int) bool {
	// TODO
	return true
}
