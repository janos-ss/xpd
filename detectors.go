package xpd

import (
	"strings"
	"regexp"
	"math"
)

type sameBodyDetector struct{}

func (detector sameBodyDetector) findDuplicates(post Post, oldPosts []Post) []Post {
	duplicates := make([]Post, 0)
	for _, oldPost := range oldPosts {
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
	limitRatio := 0.1
	limit := float64(total) * limitRatio

	duplicates := make([]Post, 0)
	for _, oldPost := range oldPosts {
		otherWordCounts, otherTotal := calcWordCounts(oldPost.Body)
		if similarEnoughCounts(total, otherTotal, limitRatio) && similarEnough(wordCounts, otherWordCounts, limit) {
			duplicates = append(duplicates, oldPost);
		}
	}
	return duplicates
}

func calcWordCounts(text string) (wordCountMap, int) {
	wordCounts := make(wordCountMap)
	total := 0
	for _, word := range splitToWords(text) {
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

func similarEnoughCounts(base, other int, limitRatio float64) bool {
	interval := applyRatio(base, limitRatio)
	return base - interval <= other && other <= base + interval
}

func applyRatio(base int, ratio float64) int {
	return int(float64(base) * ratio)
}

func similarEnough(first, second wordCountMap, limit float64) bool {
	diffs := calcWordCountDiffs(first, second) + calcWordCountDiffs(second, first)
	return diffs < limit
}

func calcWordCountDiffs(first, second wordCountMap) float64 {
	var diffs float64 = 0
	for word, count := range first {
		otherCount, ok := second[word]
		if ok {
			diffs += math.Abs(float64(count - otherCount)) / 2
		} else {
			diffs += float64(count)
		}
	}
	return diffs
}
