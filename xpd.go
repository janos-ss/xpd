package xpd

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"log"
	"time"
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
	Id  string
	Url string
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

	for _, reader := range context.readers {
		log.Printf("waitForpost for %s\n", reader.GetFeed().Id)
		go waitForPosts(reader, posts)
	}

	for {
		processQueue(context, posts)
	}
}

type Config struct {
	Feeds     []Feed `yaml:"feeds,omitempty"`
	Detectors []string `yaml:"detectors,omitempty"`
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

	log.Printf("Feeds: %#v\n", config.Feeds)
	log.Printf("Detectors: %#v\n", config.Detectors)

	readers := make([]FeedReader, 0)
	for _, feed := range config.Feeds {
		log.Printf("Creating reader for: %#v\n", feed.Id)
		readers = append(readers, NewRssReader(feed.Url, feed))
	}
	detectors := make([]Detector, 0)
	for _, detector := range config.Detectors {
		switch detector {
		case "sameBodyDetector":
			detectors = append(detectors, sameBodyDetector{})
		case "similarWordCountDetector":
			detectors = append(detectors, similarWordCountDetector{})
		default:
			panic("unrecognized detector")
		}
	}

	listeners := make([]Listener, 0)

	return Context{config.Feeds, readers, detectors, listeners, nil}
}

func waitForPosts(reader FeedReader, posts chan <- Post) {
	log.Printf("listening on feed=%s\n", reader.GetFeed().Id)
	for {
		//log.Printf("getting new posts for %s\n", reader.GetFeed().Id)
		for _, post := range reader.GetNewPosts() {
			posts <- post
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func processQueue(context Context, posts chan Post) {
	repo := context.postRepository
	recent := repo.findRecent()

	post := <-posts

	repo.add(post)

	for _, detector := range context.detectors {
		possibleDuplicates := detector.findDuplicates(post, recent)
		if len(possibleDuplicates) > 0 {
			for _, listener := range context.listeners {
				listener.onDuplicates(post, possibleDuplicates)
			}
			break
		}
	}
}

func ellipsize(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[0:length - 3] + "..."
}
