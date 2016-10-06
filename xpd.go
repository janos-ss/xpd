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

type Detector interface {
	findSimilar(Post, []Feed) []Post
}

type Listener interface {
	onSimilarDetected(Post, []Post)
}

type Context struct {
	feeds     []Feed
	readers   []FeedReader
	detectors []Detector
	listeners []Listener
}

func Run(configfile string) {
	context := readContext()

	posts := make(chan Post)

	for _, reader := range (context.readers) {
		go waitForPosts(reader, posts)
	}

	for {
		processQueue(context, posts)
	}
}

func readContext() Context {
	// TODO read configuration and create context elements
	return Context{}
}

func waitForPosts(reader FeedReader, posts chan<- Post) {
	for {
		for _, post := range (reader.GetNewPosts()) {
			posts <- post
		}
	}
}

func processQueue(context Context, posts chan Post) {
	post := <-posts
	for _, detector := range (context.detectors) {
		similar := detector.findSimilar(post, context.feeds)
		for _, listener := range (context.listeners) {
			listener.onSimilarDetected(post, similar)
		}
	}
}
