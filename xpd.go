package xpd

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"fmt"
	"log"
)

type Post struct {
	Id      string
	Url     string
	Author  string
	Subject string
	Body    string
	Feed    *Feed
}

type Feed struct {
	Id    string
	Url   string
	Posts []Post
}

type FeedReader interface {
	GetFeed() Feed
	GetNewPosts() []Post
}

type Detector interface {
	findDuplicates(Post, []Post) []Post
}

type Listener interface {
	onDuplicates(Post, []Post)
}

type PostRepository interface {
	findRecent() []Post
	add(Post)
}

type simplePostRepository struct {
	posts []Post
}

func newSimplePostRepository() *simplePostRepository {
	return &simplePostRepository{}
}

func (repo simplePostRepository) findRecent() []Post {
	return repo.posts
}

func (repo *simplePostRepository) add(post Post) {
	repo.posts = append(repo.posts, post)
}

type Context struct {
	feeds          []Feed
	readers        []FeedReader
	detectors      []Detector
	listeners      []Listener
	postRepository PostRepository
}

func Run(configfile string) {
	context := readContext(configfile)

	// TODO move this to readContext
	context.postRepository = newSimplePostRepository()
	context.listeners = append(context.listeners, consolePrinterListener{})

	posts := make(chan Post)

	for _, reader := range (context.readers) {
		go waitForPosts(reader, posts)
	}

	for {
		processQueue(context, posts)
	}
}

type Config struct {
	Feeds []Feed `yaml:"feeds,omitempty"`
}

func readContext(configfile string) Context {
	filename, _ := filepath.Abs(configfile)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		panic(err)
	}

	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Value: %#v\n", config.Feeds)

	return Context{}
}

func waitForPosts(reader FeedReader, posts chan <- Post) {
	log.Printf("listening on feed=%s\n", reader.GetFeed().Id)
	for {
		for _, post := range (reader.GetNewPosts()) {
			posts <- post
		}
	}
}

func processQueue(context Context, posts chan Post) {
	repo := context.postRepository

	post := <-posts
	log.Printf("new post: feed=%s author=%s subject=%s\n", post.Feed.Id, post.Author, post.Subject)

	for _, detector := range (context.detectors) {
		possibleDuplicates := detector.findDuplicates(post, repo.findRecent())
		repo.add(post)

		for _, listener := range (context.listeners) {
			listener.onDuplicates(post, possibleDuplicates)
		}
	}
}
