package xpd

import (
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

var nonLetters = regexp.MustCompile("[^a-z]+")

type wordCountMap struct {
	counts map[string]int
	total  int
}

func newWordCountMap(text string) *wordCountMap {
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
	wcmap := newWordCountMap(post.Body)
	limitRatio := 0.1

	duplicates := make([]Post, 0)
	for _, oldPost := range oldPosts {
		otherWordCountMap := newWordCountMap(oldPost.Body)
		if wcmap.isSimilar(otherWordCountMap, limitRatio) {
			duplicates = append(duplicates, oldPost)
		}
	}
	return duplicates
}

func (wcmap *wordCountMap) isSimilar(other *wordCountMap, limitRatio float64) bool {
	limit := float64(wcmap.total) * limitRatio
	return similarCounts(wcmap.total, other.total, limitRatio) && float64(calcWordCountDiffs(wcmap, other)) < limit
}

func splitToWords(text string) []string {
	abc := strings.TrimSpace(nonLetters.ReplaceAllString(strings.ToLower(text), " "))
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

func calcWordCountDiffs(first, second *wordCountMap) int {
	diffs := 0
	for word, count := range first.counts {
		otherCount, ok := second.counts[word]
		if ok {
			diffs += abs(count-otherCount)
		} else {
			diffs += count
		}
	}
	for word, count := range second.counts {
		_, ok := first.counts[word]
		if !ok {
			diffs += count
		}
	}
	return diffs
}

func abs(num int) int {
	if num < 0 {
		return -num
	}
	return num
}
