package xpd

import (
	"math"
	"regexp"
	"strings"
)

type SameBodyDetector struct{}

func (detector SameBodyDetector) FindDuplicates(post Post, oldPosts []Post) []Post {
	duplicates := make([]Post, 0)
	for _, oldPost := range oldPosts {
		if post.Body == oldPost.Body {
			duplicates = append(duplicates, oldPost)
		}
	}
	return duplicates
}

type SimilarWordCountDetector struct{}

type wordCountMap struct {
	counts map[string]int
	total  int
}

func createWordCountMap(text string) *wordCountMap {
	counts := make(map[string]int)
	total := 0
	for _, word := range splitToWords(text) {
		if _, ok := counts[word]; !ok {
			counts[word] = 0
		}
		counts[word]++
		total++
	}
	return &wordCountMap{counts, total}
}

func (detector SimilarWordCountDetector) FindDuplicates(post Post, oldPosts []Post) []Post {
	wordCounts := createWordCountMap(post.Body)
	limitRatio := 0.1
	limit := float64(wordCounts.total) * limitRatio

	duplicates := make([]Post, 0)
	for _, oldPost := range oldPosts {
		otherWordCounts := createWordCountMap(oldPost.Body)
		if similarCounts(wordCounts.total, otherWordCounts.total, limitRatio) && similarWordCountMaps(wordCounts, otherWordCounts, limit) {
			duplicates = append(duplicates, oldPost)
		}
	}
	return duplicates
}

func splitToWords(text string) []string {
	r := regexp.MustCompile("[^a-z]+")
	abc := strings.TrimSpace(r.ReplaceAllString(strings.ToLower(text), " "))
	return strings.Split(abc, " ")
}

func similarCounts(base, other int, limitRatio float64) bool {
	interval := calcRatio(base, limitRatio)
	return isWithinRange(other, base-interval, base+interval)
}

func calcRatio(base int, ratio float64) int {
	return int(float64(base) * ratio)
}

func isWithinRange(target, start, end int) bool {
	return start <= target && target <= end
}

func similarWordCountMaps(first, second *wordCountMap, limit float64) bool {
	diffs := calcWordCountDiffs(first, second) + calcWordCountDiffs(second, first)
	return diffs < limit
}

func calcWordCountDiffs(first, second *wordCountMap) float64 {
	var diffs float64 = 0
	for word, count := range first.counts {
		otherCount, ok := second.counts[word]
		if ok {
			diffs += math.Abs(float64(count-otherCount)) / 2
		} else {
			diffs += float64(count)
		}
	}
	return diffs
}
