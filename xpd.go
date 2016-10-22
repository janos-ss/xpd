package xpd

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
	"github.com/xpd-org/xpd/mail"
	"fmt"
)

// poll RSS feeds once per 15 minutes
const rssPollingMillis = 1000 * 60 * 15

const defaultPostRepositoryCapacity = 10000

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
	FetchNewPosts() []Post
}

type Detector interface {
	FindDuplicates(Post, []Post) []Post
}

type Listener interface {
	OnDuplicates(Post, []Post)
}

type PostRepository interface {
	FindRecent() []Post
	Add(Post)
}

type defaultPostRepository struct {
	posts    []Post
	capacity int
}

func NewPostRepository() PostRepository {
	return &defaultPostRepository{capacity: defaultPostRepositoryCapacity}
}

func (repo defaultPostRepository) FindRecent() []Post {
	return repo.posts
}

func (repo *defaultPostRepository) Add(post Post) {
	var posts []Post
	if len(repo.posts) < repo.capacity {
		posts = repo.posts
	} else {
		posts = repo.posts[1:]
	}
	repo.posts = append(posts, post)
}

type Context struct {
	Readers        []FeedReader
	Detectors      []Detector
	Listeners      []Listener
	PostRepository PostRepository
}

const maxUint = ^uint(0)
const MaxInt = int(maxUint >> 1)

func RunForever(configfile string) {
	context := NewContext(ReadConfig(configfile))
	count := MaxInt / len(context.Readers)
	run(context, count)
}

func run(context Context, count int) {
	posts := make(chan Post)

	for _, reader := range context.Readers {
		go waitForPosts(reader, posts, count)
	}

	for i := 0; i < count*len(context.Readers); i++ {
		post := <-posts
		processNewPost(context, post)
	}
}

type TypeConfig struct {
	Type      string
	Params    map[string]string
}

type Config struct {
	Feeds     []Feed
	Detectors []TypeConfig
	Listeners []TypeConfig
}

func ReadConfig(configfile string) Config {
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

	return config
}

func NewContext(config Config) Context {
	readers := parseReaders(config)

	// TODO handle errors
	detectors, _ := parseDetectors(config.Detectors)

	listeners := parseListeners(config.Listeners)

	return Context{
		Readers:        readers,
		Detectors:      detectors,
		Listeners:      listeners,
		PostRepository: NewPostRepository(),
	}
}

func parseReaders(config Config) []FeedReader {
	readers := make([]FeedReader, len(config.Feeds))
	for i, feed := range config.Feeds {
		log.Println("adding feed:", feed.Id, feed.Url)
		readers[i] = NewRssReader(feed.Url, feed)
	}
	return readers
}

func parseDetectors(items []TypeConfig) ([]Detector, error) {
	detectors := make([]Detector, len(items))
	for i, config := range items {
		var detector Detector
		switch config.Type {
		default:
			return nil, fmt.Errorf("unsupported detector type: %s", config.Type)
		case "SimilarWordCountDetector":
			detector = NewSimilarWordCountDetector(0.2)
		case "SameBodyDetector":
			detector = SameBodyDetector{}
		}
		detectors[i] = detector
	}
	return detectors, nil
}

func parseListeners(items []TypeConfig) []Listener {
	listeners := make([]Listener, 1 + len(items))
	listeners[0] = ConsolePrinterListener{}

	for i, config := range items {
		var listener Listener
		switch config.Type {
		default:
			panic("unknown listener type")
		case "gmail":
			listener = MailerListener{
				Mailer: mail.GmailMailer{
					From: config.Params["from"],
					Pass: config.Params["pass"],
					Recipient: config.Params["recipient"],
					Subject: config.Params["subject"],
				},
			}
		}
		listeners[i + 1] = listener
	}
	return listeners
}

func waitForPosts(reader FeedReader, posts chan<- Post, count int) {
	log.Println("listening on feed:", reader.GetFeed().Id)
	for i := 0; i < count; i++ {
		for _, post := range reader.FetchNewPosts() {
			posts <- post
		}
		time.Sleep(rssPollingMillis * time.Millisecond)
	}
}

func processNewPost(context Context, post Post) {
	repo := context.PostRepository
	recent := repo.FindRecent()

	for _, detector := range context.Detectors {
		possibleDuplicates := detector.FindDuplicates(post, recent)
		if len(possibleDuplicates) > 0 {
			for _, listener := range context.Listeners {
				// TODO add Detector ref as param
				listener.OnDuplicates(post, possibleDuplicates)
			}
			break
		}
	}

	repo.Add(post)
}
