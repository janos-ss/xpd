package xpd

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"fmt"
)

type Post struct {
	Id      string
	Url     string
	Author  string
	Subject string
	Body    string
}

type Feed struct {
	Posts []Post
}

type FeedReader interface {
	GetNewPosts([]Post) []Post
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
	feeds     []Feed
	readers   []FeedReader
	detectors []Detector
	listeners []Listener
}

func Run(configfile string) {
	context := readContext(configfile)

	posts := make(chan Post)

	for _, reader := range (context.readers) {
		go waitForPosts(reader, posts)
	}

	for {
		processQueue(context, posts)
	}
}

type Config struct {
	Feeds []string `yaml:"feeds,omitempty"`
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
	for {
		for _, post := range (reader.GetNewPosts(make([]Post, 0))) {
			posts <- post
		}
	}
}

func processQueue(context Context, posts chan Post) {
	repo := newSimplePostRepository()

	post := <-posts
	for _, detector := range (context.detectors) {
		possibleDuplicates := detector.findDuplicates(post, repo.findRecent())
		repo.add(post)

		for _, listener := range (context.listeners) {
			listener.onDuplicates(post, possibleDuplicates)
		}
	}
}
