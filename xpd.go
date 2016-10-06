package xpd

type Post struct {
	Url     string
	Author  string
	Subject string
	Body    string
}

type Feed struct {
	Posts []Post
}

type FeedReader interface {
	GetNewPosts() []Post
}
