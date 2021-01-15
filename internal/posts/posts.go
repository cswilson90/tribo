package posts

import (
	"path/filepath"
	"io/ioutil"
	"regexp"

	log "github.com/sirupsen/logrus"
)

var metaDataMatch = regexp.MustCompile(`^metadata\.(json|ya?ml)$`)

type DirSet map[string]bool

type Posts []*Post

type Post struct {
	dir string
}

func FindPosts(baseDir string) Posts {
	log.Infof("Looking for posts recursivly in '%v'", baseDir)

	posts := findPosts(baseDir, make(DirSet))

	log.Infof("Found %v posts in '%v'", len(posts), baseDir)
	return posts
}

func findPosts(baseDir string, seenDirs DirSet) Posts {
	log.Debugf("Looking for posts in '%v'", baseDir)

	// Stop if we've seen this directory before (prevents infinite loops due to circular references)
	if seenDirs[baseDir] {
		return Posts{}
	}
	seenDirs[baseDir] = true

	fileList, err := ioutil.ReadDir(baseDir)
	if err != nil {
		log.Warnf("Could not list files in directory '%v':"+err.Error(), baseDir)
		return Posts{}
	}

	// If the directory contains a metadata file it's a post directory
	for _, file := range fileList {
		if metaDataMatch.MatchString(file.Name()) {
			return Posts{&Post{baseDir}}
		}
	}

	// Recursively check sub-direcotories for posts
	posts := make(Posts, 0)
	for _, file := range fileList {
		if file.IsDir() {
			posts = append(posts, findPosts(filepath.Join(baseDir, file.Name()), seenDirs)...)
		}
	}

	log.Debugf("Found %v posts in '%v'", len(posts), baseDir)
	return posts
}
