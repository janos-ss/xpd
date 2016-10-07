package xpd

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
