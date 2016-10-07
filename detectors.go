package xpd

import (
	"strings"
	"regexp"
	"math"
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
	limit := float64(total) / 10

	duplicates := make([]Post, 0)
	for _, oldPost := range (oldPosts) {
		otherWordCounts, otherTotal := calcWordCounts(oldPost.Body)
		if similarEnoughCounts(total, otherTotal) && similarEnough(wordCounts, otherWordCounts, limit) {
			duplicates = append(duplicates, oldPost);
		}
	}
	return duplicates
}

func calcWordCounts(text string) (wordCountMap, int) {
	wordCounts := make(wordCountMap)
	total := 0
	for _, word := range (splitToWords(text)) {
		if _, ok := wordCounts[word]; !ok {
			wordCounts[word] = 0
		}
		wordCounts[word]++
		total++
	}
	return wordCounts, total
}

func splitToWords(text string) []string {
	r := regexp.MustCompile("[^a-z]+")
	abc := strings.TrimSpace(r.ReplaceAllString(strings.ToLower(text), " "))
	return strings.Split(abc, " ")
}

func similarEnoughCounts(a, b int) bool {
	ratio := float64(a) / float64(b)
	interval := .1
	return 1 - interval < ratio && ratio < 1 + interval
}

func similarEnough(first, second wordCountMap, limit float64) bool {
	diffs := calcWordCountDiffs(first, second) + calcWordCountDiffs(second, first)
	return diffs < limit
}

func calcWordCountDiffs(first, second wordCountMap) float64 {
	var diffs float64 = 0
	for word, count := range(first) {
		otherCount, ok := second[word]
		if ok {
			diffs += math.Abs(float64(count - otherCount)) / 2
		} else {
			diffs += float64(count)
		}
	}
	return diffs
}
