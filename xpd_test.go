package xpd

import (
	"testing"
)

func addPostToRepo(repo PostRepository, post Post) {
	repo.add(post)
}

func Test_adding_to_repo(t*testing.T) {
	repo := NewSimplePostRepository()
	addPostToRepo(repo, Post{})

	if len(repo.findRecent()) == 0 {
		t.Errorf("PostRepository should not be empty after post added")
	}
}
