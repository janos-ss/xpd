package xpd

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"log"
	"time"
	"reflect"
)

// poll RSS feeds once per 15 minutes
const rss_polling_millis = 1000 * 60 * 15

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

type DetectorRegistry interface {
	Register(Detector)
	Get(string) Detector
}

type defaultDetectorRegistry struct {
	detectors map[string]Detector
}

func NewDetectorRegistry() DetectorRegistry {
	return defaultDetectorRegistry{make(map[string]Detector)}
}

func (reg defaultDetectorRegistry) Register(detector Detector) {
	name := reflect.TypeOf(detector).Name()
	reg.detectors[name] = detector
}

func (reg defaultDetectorRegistry) Get(name string) Detector {
	if detector, ok := reg.detectors[name]; ok {
		return detector
	}
	panic("no such detector: " + name)
}

type Listener interface {
	OnDuplicates(Post, []Post)
}

type PostRepository interface {
	FindRecent() []Post
	Add(Post)
}

type defaultPostRepository struct {
	posts []Post
}

func NewPostRepository() PostRepository {
	return &defaultPostRepository{}
}

func (repo defaultPostRepository) FindRecent() []Post {
	return repo.posts
}

func (repo *defaultPostRepository) Add(post Post) {
	repo.posts = append(repo.posts, post)
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
	context := CreateContext(ReadConfig(configfile))
	count := MaxInt / len(context.Readers)
	run(context, count)
}

func run(context Context, count int) {
	posts := make(chan Post)

	for _, reader := range context.Readers {
		go waitForPosts(reader, posts, count)
	}

	for i := 0; i < count * len(context.Readers); i++ {
		post := <-posts
		processNewPost(context, post)
	}
}

type Config struct {
	Feeds         []Feed `yaml:"feeds,omitempty"`
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

func CreateContext(config Config) Context {
	readers := getReaders(config)

	detectorRegistry := NewDetectorRegistry()
	detectorRegistry.Register(SameBodyDetector{})
	detectorRegistry.Register(SimilarWordCountDetector{})

	detectors := getDetectors(detectorRegistry, config.DetectorNames)

	listeners := []Listener{ConsolePrinterListener{}}

	return Context{
		Readers: readers,
		Detectors: detectors,
		Listeners: listeners,
		PostRepository: NewPostRepository(),
	}
}

func getReaders(config Config) []FeedReader {
	readers := make([]FeedReader, 0)
	for _, feed := range config.Feeds {
		log.Printf("Creating reader for: %#v\n", feed.Id)
		readers = append(readers, NewRssReader(feed.Url, feed))
	}
	return readers
}

func getDetectors(reg DetectorRegistry, detectorNames []string) []Detector {
	detectors := make([]Detector, 0)
	for _, name := range detectorNames {
		detectors = append(detectors, reg.Get(name))
	}
	return detectors
}

func waitForPosts(reader FeedReader, posts chan <- Post, count int) {
	log.Printf("listening on feed=%s\n", reader.GetFeed().Id)
	for i := 0; i < count; i++ {
		for _, post := range reader.FetchNewPosts() {
			posts <- post
		}
		time.Sleep(rss_polling_millis * time.Millisecond)
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
	return s[0:length - 3] + "..."
}
