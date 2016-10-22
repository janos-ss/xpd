package xpd

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"time"
	"github.com/xpd-org/xpd/mail"
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

type DetectorRegistry struct {
	detectors map[string]Detector
}

func NewDetectorRegistry() DetectorRegistry {
	return DetectorRegistry{make(map[string]Detector)}
}

func (registry DetectorRegistry) Register(detector Detector) {
	name := reflect.TypeOf(detector).String()
	if reflect.TypeOf(detector).Kind() == reflect.Ptr {
		name = name[1:]
	}
	log.Println("adding detector:", name)
	registry.detectors[name] = detector
}

func (registry DetectorRegistry) Get(name string) (Detector, bool) {
	detector, ok := registry.detectors[name]
	return detector, ok
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

type ListenerConfig struct {
	Type      string
	Params    map[string]string
}

type Config struct {
	Feeds     []Feed
	Detectors []string
	Listeners []ListenerConfig
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

	detectorRegistry := getDetectorRegistry()

	detectors := parseDetectors(detectorRegistry, config.Detectors)

	listeners := parseListeners(config)

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

func getDetectorRegistry() DetectorRegistry {
	registry := NewDetectorRegistry()
	registry.Register(SameBodyDetector{})
	registry.Register(NewSimilarWordCountDetector(0.2))

	return registry
}

func parseDetectors(registry DetectorRegistry, names []string) []Detector {
	detectors := make([]Detector, 0)
	for _, name := range names {
		if detector, ok := registry.Get(name); ok {
			detectors = append(detectors, detector)
		} else {
			log.Println("no such detector:", name)
		}
	}
	return detectors
}

func parseListeners(config Config) []Listener {
	listeners := make([]Listener, 1 + len(config.Listeners))
	listeners[0] = ConsolePrinterListener{}

	for i, config := range config.Listeners {
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
