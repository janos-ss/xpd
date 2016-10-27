package xpd

import (
	"reflect"
	"strconv"
	"testing"
)

func Test_defaultPostRepository_should_cycle_posts_to_keep_capacity(t *testing.T) {
	post1 := Post{Id: "1"}
	post2 := Post{Id: "2"}

	var repo PostRepository = &defaultPostRepository{capacity: 2}
	repo.Add(post1)
	repo.Add(post2)

	if expected, actual := []Post{post1, post2}, repo.FindRecent(); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("got %#v; expected %#v", actual, expected)
	}

	post3 := Post{Id: "3"}
	repo.Add(post3)
	if expected, actual := []Post{post2, post3}, repo.FindRecent(); !reflect.DeepEqual(actual, expected) {
		t.Fatalf("got %#v; expected %#v", actual, expected)
	}
}

func assertPanic(t *testing.T, message string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf(message)
		}
	}()
	f()
}

func Test_ParseContext(t *testing.T) {
	config := &Config{
		Feeds:     []Feed{{}},
		Detectors: []TypeConfig{{Type: "SameBodyDetector"}},
	}
	context, err := ParseContext(config)
	if err != nil {
		t.Fatal("failed to parse context from valid config")
	}

	if len(context.Readers) != len(config.Feeds) {
		t.Errorf("got different number of feed readers than specified feeds; %#v <- %#v", context.Readers, config.Feeds)
	}
	if len(context.Readers) < 1 {
		t.Error("got no feed readers, expected at least 1")
	}

	if len(context.Detectors) != len(config.Detectors) {
		t.Errorf("got different number of detectors than specified names; %#v <- %#v", context.Detectors, config.Detectors)
	}
	if len(context.Detectors) < 1 {
		t.Error("got no detectors, expected at least 1")
	}

	if len(context.Listeners) < 1 {
		t.Error("got no listeners, expected at least 1")
	}

	if context.PostRepository == nil {
		t.Error("got nil PostRepository, expected non-nil")
	}

	brokenConfig := *config
	brokenConfig.Feeds = nil
	if _, err := ParseContext(&brokenConfig); err == nil {
		t.Error("config parser should fail if no feeds")
	}

	brokenConfig = *config
	brokenConfig.Detectors = nil
	if _, err := ParseContext(&brokenConfig); err == nil {
		t.Error("config parser should fail if no detectors")
	}

	brokenConfig.Detectors = []TypeConfig{{Type: "nonexistent"}}
	if _, err := ParseContext(&brokenConfig); err == nil {
		t.Error("config parser should fail with invalid detector")
	}

	brokenConfig = *config
	brokenConfig.Listeners = []TypeConfig{{Type: "nonexistent"}}
	if _, err := ParseContext(&brokenConfig); err == nil {
		t.Error("config parser should fail with invalid listener")
	}
}

func Test_parseListeners_gmail(t *testing.T) {
	listeners, err := parseListeners([]TypeConfig{{Type: "gmail"}})
	if err != nil {
		t.Fatalf("got error: %s; expected successful parsing of gmail sender listener", err)
	}
	if len(listeners) != 1 {
		t.Fatalf("got %d listeners; expected 1", len(listeners))
	}
}

func Test_ParseConfig_valid_example(t *testing.T) {
	config, err := ParseConfig("xpd.yml.example")
	if err != nil {
		t.Fatal(err)
	}

	if len(config.Feeds) < 1 {
		t.Error("got no feeds, expected at least 1")
	}
	if len(config.Detectors) < 1 {
		t.Error("got no detectors, expected at least 1")
	}
}

func Test_ParseConfig_nonexistent_should_crash(t *testing.T) {
	if _, err := ParseConfig("nonexistent"); err == nil {
		t.Fatal("should fail to read config of non-existent file")
	}
}

func Test_ParseConfig_malformed_should_crash(t *testing.T) {
	if _, err := ParseConfig("xpd.go"); err == nil {
		t.Fatal("should fail to parse malformed config file")
	}
}

func Test_parseDetectors_SimilarWordCountDetector_with_default_maxDiffRatio(t *testing.T) {
	config := &Config{
		Feeds:     []Feed{{}},
		Detectors: []TypeConfig{{Type: "SimilarWordCountDetector"}},
	}

	context, err := ParseContext(config)
	if err != nil {
		t.Fatalf("got %#v; expected nil (successful parsing of configuration)", err)
	}
	if len(context.Detectors) != 1 {
		t.Fatalf("got %d detectors; expected 1", len(context.Detectors))
	}

	defaultMaxDiffRatio := 0.1
	detector := context.Detectors[0].(SimilarWordCountDetector)
	if detector.maxDiffRatio != defaultMaxDiffRatio {
		t.Fatalf("got maxDiffRatio=%f; expected %f", detector.maxDiffRatio, defaultMaxDiffRatio)
	}
}

func Test_parseDetectors_SimilarWordCountDetector_with_custom_maxDiffRatio(t *testing.T) {
	config := &Config{
		Feeds:     []Feed{{}},
		Detectors: []TypeConfig{{Type: "SimilarWordCountDetector"}},
	}

	customMaxDiffRatio := 0.2
	config.Detectors[0].Params = make(map[string]string)
	config.Detectors[0].Params["maxDiffRatio"] = strconv.FormatFloat(customMaxDiffRatio, 'f', -1, 64)

	context, err := ParseContext(config)
	if err != nil {
		t.Fatalf("got %#v; expected nil (successful parsing of configuration)", err)
	}
	if len(context.Detectors) != 1 {
		t.Fatalf("got %d detectors; expected 1", len(context.Detectors))
	}

	detector := context.Detectors[0].(SimilarWordCountDetector)
	if detector.maxDiffRatio != customMaxDiffRatio {
		t.Fatalf("got maxDiffRatio=%f; expected %f", detector.maxDiffRatio, customMaxDiffRatio)
	}
}

func Test_parseDetectors_SimilarWordCountDetector_with_malformed_maxDiffRatio(t *testing.T) {
	config := &Config{
		Feeds:     []Feed{{}},
		Detectors: []TypeConfig{{Type: "SimilarWordCountDetector"}},
	}

	config.Detectors[0].Params = make(map[string]string)
	config.Detectors[0].Params["maxDiffRatio"] = "malformed"

	_, err := ParseContext(config)
	if err == nil {
		t.Fatal("got success; expected parsing to fail due to malformed maxDiffRatio param")
	}
}

type mockListener struct {
	invoked bool
}

func (listener *mockListener) OnDuplicates(Post, []Post) {
	listener.invoked = true
}

func Test_processPost(t *testing.T) {
	post := Post{}

	listener := &mockListener{}
	repo := NewPostRepository()

	context := &Context{
		Detectors:      []Detector{SameBodyDetector{}},
		Listeners:      []Listener{listener},
		PostRepository: repo,
	}

	processNewPost(context, post)
	if listener.invoked {
		t.Error("mock listener was invoked, but should not have been")
	}
	if len(repo.FindRecent()) != 1 {
		t.Fatal("got != 1 recent posts, expected one dummy post added")
	}

	processNewPost(context, post)
	if !listener.invoked {
		t.Error("mock listener should have been invoked, but it was not")
	}
	if len(repo.FindRecent()) != 2 {
		t.Fatal("got != 2 recent posts, expected the dummy post added twice")
	}
}

type mockReader struct {
	post Post
}

func (reader *mockReader) GetFeed() Feed {
	return Feed{Id: "dummy"}
}

func (reader *mockReader) FetchNewPosts() []Post {
	return []Post{{}}
}

func Test_waitForPosts(t *testing.T) {
	post := Post{}

	reader := &mockReader{post: post}
	posts := make(chan Post)

	go waitForPosts(reader, posts, 1)

	if received := <-posts; received != post {
		t.Fatalf("got %#v, expected %#v", received, post)
	}
}

func restoreDefaultCount() {
	defaultCount = getDefaultCount()
}

func Test_run(t *testing.T) {
	defaultCount = 0
	defer restoreDefaultCount()

	post := Post{}

	reader := &mockReader{post: post}
	listener := &mockListener{}
	repo := NewPostRepository()

	context := &Context{
		Readers:        []FeedReader{reader},
		Detectors:      []Detector{SameBodyDetector{}},
		Listeners:      []Listener{listener},
		PostRepository: repo,
	}

	run(context, 1)

	if !reflect.DeepEqual([]Post{post}, repo.FindRecent()) {
		t.Fatalf("got %#v, expected []Post{%#v}", repo.FindRecent(), post)
	}
}

func Test_RunForever_fails_if_config_file_nonexistent(t *testing.T) {
	defaultCount = 0
	defer restoreDefaultCount()

	if RunForever("xpd.yml.example") != nil {
		t.Fatal("got failure; expected RunForever to succeed with valid config file")
	}

	if RunForever("nonexistent") == nil {
		t.Fatal("got success; expected RunForever to fail if config file nonexistent")
	}
}

func Test_runForever_fails_if_config_invalid(t *testing.T) {
	defaultCount = 0
	defer restoreDefaultCount()

	validConfig := &Config{
		Feeds:     []Feed{{}},
		Detectors: []TypeConfig{{Type: "SimilarWordCountDetector"}},
	}

	if err := runForever(validConfig); err != nil {
		t.Fatalf("got failure: %s; runForever should have worked with valid config", err)
	}

	var brokenConfig Config

	brokenConfig = *validConfig
	brokenConfig.Feeds = nil
	if runForever(&brokenConfig) == nil {
		t.Fatal("got success; expected runForever to fail if feeds missing")
	}

	brokenConfig = *validConfig
	brokenConfig.Detectors = nil
	if runForever(&brokenConfig) == nil {
		t.Fatal("got success; expected runForever to fail if detectors missing")
	}
}

func Test_defaultCount(t *testing.T) {
	if expected := 1<<63 - 1; getDefaultCount() != expected {
		t.Fatalf("got defaultCount = %d; expected %d", getDefaultCount(), expected)
	}
}
