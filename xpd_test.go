package xpd

import (
	"reflect"
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

func Test_DetectorRegistry(t *testing.T) {
	detector := NewSimilarWordCountDetector(0.1)

	reg := NewDetectorRegistry()
	reg.Register(detector)

	if d, ok := reg.Get("xpd.SimilarWordCountDetector"); !ok || !reflect.DeepEqual(d, detector) {
		t.Errorf("got %#v, expected %#v", d, detector)
	}

	if d, ok := reg.Get("nonexistent"); ok {
		t.Fatalf("got %#v, but expected no such detector", d)
	}
}

func Test_parseDetectors_should_work_transparently_for_both_values_and_pointers(t *testing.T) {
	expected := []Detector{SameBodyDetector{}, &SimilarWordCountDetector{}}

	reg := NewDetectorRegistry()
	for _, detector := range expected {
		reg.Register(detector)
	}

	detectors := parseDetectors(reg, []string{"xpd.SameBodyDetector", "xpd.SimilarWordCountDetector", "garbage"})

	if !reflect.DeepEqual(detectors, expected) {
		t.Errorf("got %#v, expected %#v", detectors, expected)
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

func Test_NewContext(t *testing.T) {
	config := Config{
		Feeds: []Feed{
			{Id: "dummy1", Url: "dummy1"},
			{Id: "dummy2", Url: "dummy2"},
		},
		Detectors: []string{"xpd.SameBodyDetector"},
		Listeners: []ListenerConfig{{Type: "console"}},
	}
	context := NewContext(config)

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
}

func Test_ReadConfig_valid_example(t *testing.T) {
	config := ReadConfig("xpd.yml.example")

	if len(config.Feeds) < 1 {
		t.Error("got no feeds, expected at least 1")
	}
	if len(config.Detectors) < 1 {
		t.Error("got no detectors, expected at least 1")
	}
}

func Test_ReadConfig_nonexistent_should_crash(t *testing.T) {
	assertPanic(t, "did not crash on non-existent config file, but it should have", func() {
		ReadConfig("nonexistent")
	})
}

func Test_ReadConfig_malformed_should_crash(t *testing.T) {
	assertPanic(t, "did not crash on malformed config file, but it should have", func() {
		ReadConfig("xpd.go")
	})
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

	context := Context{
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

func Test_run(t *testing.T) {
	post := Post{}

	reader := &mockReader{post: post}
	listener := &mockListener{}
	repo := NewPostRepository()

	context := Context{
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
