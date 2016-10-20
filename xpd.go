package xpd

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"time"
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
	log.Print("adding detector:", name)
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

type Config struct {
	Feeds         []Feed   `yaml:"feeds,omitempty"`
	DetectorNames []string `yaml:"detectors,omitempty"`
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

	log.Printf("Feeds: %#v\n", config.Feeds)
	log.Printf("Detectors: %#v\n", config.DetectorNames)

	return config
}

func NewContext(config Config) Context {
	readers := getReaders(config)

	detectorRegistry := getDetectorRegistry()

	detectors := getDetectors(detectorRegistry, config.DetectorNames)

	listeners := []Listener{ConsolePrinterListener{}}

	return Context{
		Readers:        readers,
		Detectors:      detectors,
		Listeners:      listeners,
		PostRepository: NewPostRepository(),
	}
}

func getDetectorRegistry() DetectorRegistry {
	registry := NewDetectorRegistry()
	registry.Register(SameBodyDetector{})
	registry.Register(NewSimilarWordCountDetector(0.2))

	return registry
}

func getReaders(config Config) []FeedReader {
	readers := make([]FeedReader, len(config.Feeds))
	for i, feed := range config.Feeds {
		log.Printf("adding reader for: %#v\n", feed.Id)
		readers[i] = NewRssReader(feed.Url, feed)
	}
	return readers
}

func getDetectors(registry DetectorRegistry, detectorNames []string) []Detector {
	detectors := make([]Detector, 0)
	for _, name := range detectorNames {
		if detector, ok := registry.Get(name); ok {
			detectors = append(detectors, detector)
		} else {
			log.Printf("no such detector: %s", name)
		}
	}
	return detectors
}

func waitForPosts(reader FeedReader, posts chan<- Post, count int) {
	log.Printf("listening on feed=%s\n", reader.GetFeed().Id)
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

func ellipsize(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[0:length-3] + "..."
}
