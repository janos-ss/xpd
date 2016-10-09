package xpd

import (
	"gopkg.in/yaml.v2"
	"path/filepath"
	"io/ioutil"
	"log"
	"time"
	"reflect"
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
	FetchNewPosts() []Post
}

type Detector interface {
	findDuplicates(Post, []Post) []Post
}

type DetectorRegistry interface {
	register(Detector)
	get(string) Detector
}

type simpleDetectorRegistry struct {
	detectors map[string]Detector
}

func newSimpleDetectorRegistry() DetectorRegistry {
	return simpleDetectorRegistry{make(map[string]Detector)}
}

func (reg simpleDetectorRegistry) register(detector Detector) {
	name := reflect.TypeOf(detector).Name()
	reg.detectors[name] = detector
}

func (reg simpleDetectorRegistry) get(name string) Detector {
	if detector, ok := reg.detectors[name]; ok {
		return detector
	}
	panic("no such detector: " + name)
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
	readers        []FeedReader
	detectors      []Detector
	listeners      []Listener
	postRepository PostRepository
}

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)

func Run(configfile string) {
	run(createContext(readConfig(configfile)), MaxInt)
}

func run(context Context, count int) {
	posts := make(chan Post)

	for _, reader := range context.readers {
		go waitForPosts(reader, posts, count)
	}

	for i := 0; i < count * len(context.readers); i++ {
		post := <-posts
		processNewPost(context, post)
	}
}

type Config struct {
	Feeds         []Feed `yaml:"feeds,omitempty"`
	DetectorNames []string `yaml:"detectors,omitempty"`
}

func readConfig(configfile string) Config {
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

func createContext(config Config) Context {
	readers := make([]FeedReader, 0)
	for _, feed := range config.Feeds {
		log.Printf("Creating reader for: %#v\n", feed.Id)
		readers = append(readers, NewRssReader(feed.Url, feed))
	}

	detectorRegistry := newSimpleDetectorRegistry()
	detectorRegistry.register(sameBodyDetector{})
	detectorRegistry.register(similarWordCountDetector{})

	detectors := getDetectors(detectorRegistry, config.DetectorNames)

	listeners := []Listener{consolePrinterListener{}}

	return Context{
		readers: readers,
		detectors: detectors,
		listeners: listeners,
		postRepository: newSimplePostRepository(),
	}
}

func getDetectors(reg DetectorRegistry, detectorNames []string) []Detector {
	detectors := make([]Detector, 0)
	for _, name := range detectorNames {
		detectors = append(detectors, reg.get(name))
	}
	return detectors
}

func waitForPosts(reader FeedReader, posts chan <- Post, count int) {
	log.Printf("listening on feed=%s\n", reader.GetFeed().Id)
	for i := 0; i < count; i++ {
		for _, post := range reader.FetchNewPosts() {
			posts <- post
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func processNewPost(context Context, post Post) {
	repo := context.postRepository
	recent := repo.findRecent()

	for _, detector := range context.detectors {
		possibleDuplicates := detector.findDuplicates(post, recent)
		if len(possibleDuplicates) > 0 {
			for _, listener := range context.listeners {
				// TODO add Detector ref as param
				listener.onDuplicates(post, possibleDuplicates)
			}
			break
		}
	}

	repo.add(post)
}

func ellipsize(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[0:length - 3] + "..."
}
