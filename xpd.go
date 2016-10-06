package xpd

type Post struct {
	author string
	subject string
	body string
}

type Feed struct {
	posts []Post
}

type FeedReader interface {
	GetNewPosts() []Post
}

type MboxReader struct {
	feed Feed
}

func (reader MboxReader) GetNewPosts() []Post {
	// read posts from source
	// for each new post
	// 	add to feed
	// 	add to new

	return []Post{}
}

func demo() {
	feed := Feed{}
	reader := MboxReader{feed}

	reader.GetNewPosts()
}
