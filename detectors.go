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

type wordCountCache map[string]wordCountMap

type SimilarWordCountDetector struct {
	maxDiffRatio float64
	indexMap     wordCountCache
}

func NewSimilarWordCountDetector(maxDiffRatio float64) SimilarWordCountDetector {
	return SimilarWordCountDetector{
		maxDiffRatio: maxDiffRatio,
		indexMap:     make(wordCountCache),
	}
}

var nonLetters = regexp.MustCompile("[^a-z]+")

type wordCountMap struct {
	counts map[string]int
	total  int
}

func newWordCountMap(text string) wordCountMap {
	counts := make(map[string]int)
	total := 0
	for _, word := range splitToWords(text) {
		if _, ok := counts[word]; !ok {
			counts[word] = 0
		}
		counts[word]++
		total++
	}
	return wordCountMap{counts, total}
}

func splitToWords(text string) []string {
	abc := strings.TrimSpace(nonLetters.ReplaceAllString(strings.ToLower(text), " "))
	return strings.Split(abc, " ")
}

func (detector SimilarWordCountDetector) getWordCountMap(post Post) wordCountMap {
	if wcmap, ok := detector.indexMap[post.Id]; ok {
		return wcmap
	}
	wcmap := newWordCountMap(post.Body)
	detector.indexMap[post.Id] = wcmap
	return wcmap
}

func (detector SimilarWordCountDetector) FindDuplicates(post Post, oldPosts []Post) []Post {
	wcmap := detector.getWordCountMap(post)
	defer detector.cleanIndex(append(oldPosts, post))

	duplicates := make([]Post, 0)
	for _, oldPost := range oldPosts {
		otherWordCountMap := detector.getWordCountMap(oldPost)
		if wcmap.isSimilar(otherWordCountMap, detector.maxDiffRatio) {
			duplicates = append(duplicates, oldPost)
		}
	}
	return duplicates
}

func (detector SimilarWordCountDetector) cleanIndex(posts []Post) {
	if len(detector.indexMap) <= len(posts) {
		return
	}

	seen := make(map[string]bool)
	for _, post := range posts {
		seen[post.Id] = true
	}

	for key := range detector.indexMap {
		if !seen[key] {
			delete(detector.indexMap, key)
		}
	}
}

func (wcmap wordCountMap) isSimilar(other wordCountMap, limitRatio float64) bool {
	limit := float64(wcmap.total) * limitRatio
	return float64(abs(wcmap.total-other.total)) < limit && float64(calcWordCountDiffs(wcmap, other)) < limit
}

func calcWordCountDiffs(first, second wordCountMap) int {
	diffs := 0
	for word, count := range first.counts {
		otherCount, ok := second.counts[word]
		if ok {
			diffs += abs(count - otherCount)
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
