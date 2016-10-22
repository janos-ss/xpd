package xpd

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
	"github.com/xpd-org/xpd/mail"
	"fmt"
	"errors"
)

// poll RSS feeds once per 15 minutes
const rssPollingMillis = 1000 * 60 * 15

// number of recent posts to keep in memory
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

type TypeConfig struct {
	Type      string
	Params    map[string]string
}

type Config struct {
	Feeds     []Feed
	Detectors []TypeConfig
	Listeners []TypeConfig
}

func ReadConfig(path string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

type Context struct {
	Readers        []FeedReader
	Detectors      []Detector
	Listeners      []Listener
	PostRepository PostRepository
}

func ParseContext(config *Config) (*Context, error) {
	readers := parseReaders(config)

	if len(config.Feeds) == 0 {
		return nil, errors.New("configuration error: you must configure at least one feed")
	}

	detectors, err := parseDetectors(config.Detectors)
	if err != nil {
		return nil, err
	}

	if len(config.Detectors) == 0 {
		return nil, errors.New("configuration error: you must configure at least one detector")
	}

	extraListeners, err := parseListeners(config.Listeners)
	if err != nil {
		return nil, err
	}

	listeners := []Listener{ConsolePrinterListener{}}
	listeners = append(listeners, extraListeners...)

	context := &Context{
		Readers:        readers,
		Detectors:      detectors,
		Listeners:      listeners,
		PostRepository: NewPostRepository(),
	}
	return context, nil
}

func parseReaders(config *Config) []FeedReader {
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

func parseListeners(items []TypeConfig) ([]Listener, error) {
	listeners := make([]Listener, len(items))
	for i, config := range items {
		var listener Listener
		switch config.Type {
		default:
			return nil, fmt.Errorf("unsupported listener type: %s", config.Type)
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
		listeners[i] = listener
	}
	return listeners, nil
}

// the default number of posts to read; normally infinity, set to 0 by some tests
var defaultCount int

func init() {
	defaultCount = getDefaultCount()
}

func getDefaultCount() int {
	maxUint := ^uint(0)
	maxInt := int(maxUint >> 1)
	return maxInt
}

func RunForever(path string) error {
	config, err := ReadConfig(path)
	if err != nil {
		return err
	}

	return runForever(config)
}

func runForever(config *Config) error {
	context, err := ParseContext(config)
	if err != nil {
		return err
	}

	run(context, defaultCount)

	return nil
}

func run(context *Context, count int) {
	posts := make(chan Post)

	for _, reader := range context.Readers {
		go waitForPosts(reader, posts, count)
	}

	for i := 0; i < count; i++ {
		post := <-posts
		processNewPost(context, post)
	}
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

func processNewPost(context *Context, post Post) {
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
