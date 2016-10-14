package xpd

import (
	"reflect"
	"testing"
)

func Test_adding_to_repo(t *testing.T) {
	var repo PostRepository = NewPostRepository()
	repo.Add(Post{})

	if len(repo.FindRecent()) == 0 {
		t.Errorf("PostRepository should not be empty after post added")
	}
}

func Test_DefaultDetectorRegistry(t *testing.T) {
	detector := SimilarWordCountDetector{}

	reg := NewDetectorRegistry()
	reg.Register(detector)

	if d := reg.Get("SimilarWordCountDetector"); d != detector {
		t.Errorf("got %#v, expected %#v", d, detector)
	}

	assertPanic(t, "did not crash on unknown Detector, but it should have", func() {
		reg.Get("nonexistent")
	})
}

func Test_getDetectors(t *testing.T) {
	reg := NewDetectorRegistry()
	reg.Register(SameBodyDetector{})
	reg.Register(SimilarWordCountDetector{})

	detectors := getDetectors(reg, []string{"SameBodyDetector", "SimilarWordCountDetector"})
	expected := []Detector{SameBodyDetector{}, SimilarWordCountDetector{}}

	if !reflect.DeepEqual(detectors, expected) {
		t.Errorf("got %#v, expected %#v", detectors, expected)
	}

	assertPanic(t, "did not crash on unknown Detector, but it should have", func() {
		getDetectors(reg, []string{"nonexistent"})
	})
}

func assertPanic(t *testing.T, message string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf(message)
		}
	}()
	f()
}

func Test_CreateContext(t *testing.T) {
	config := Config{
		Feeds: []Feed{
			{Id: "dummy1", Url: "dummy1"},
			{Id: "dummy2", Url: "dummy2"},
		},
		DetectorNames: []string{"SameBodyDetector"},
	}
	context := CreateContext(config)

	if len(context.Readers) != len(config.Feeds) {
		t.Errorf("got different number of feed readers than specified feeds; %#v <- %#v", context.Readers, config.Feeds)
	}
	if len(context.Readers) < 1 {
		t.Error("got no feed readers, expected at least 1")
	}

	if len(context.Detectors) != len(config.DetectorNames) {
		t.Errorf("got different number of detectors than specified names; %#v <- %#v", context.Detectors, config.DetectorNames)
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
	if len(config.DetectorNames) < 1 {
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

func Test_ellipsize_someString_15_is_someString(t *testing.T) {
	s := "someString"
	if actual := ellipsize(s, 15); actual != s {
		t.Fatalf("got %s; expected %s", actual, s)
	}
}

func Test_ellipsize_someString_7_is_somedots(t *testing.T) {
	s := "someString"
	if actual, expected := ellipsize(s, 7), "some..."; actual != expected {
		t.Fatalf("got %s; expected %s", actual, expected)
	}
}
