package posts

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"

	log "github.com/sirupsen/logrus"
)

var metaDataMatch = regexp.MustCompile(`^metadata\.(json|ya?ml)$`)

type DirSet map[string]bool

type Posts []*Post

type Post struct {
	dir string
}

// FindPosts recursively searches a directory for posts.
func FindPosts(baseDir string) Posts {
	log.Infof("Looking for posts recursivly in '%v'", baseDir)

	toProcess := list.New()
	toProcess.PushBack(baseDir)

	posts := make(Posts, 0)

	for toProcess.Len() > 0 {
		nextElement := toProcess.Front()
		nextDir := nextElement.Value.(string)
		toProcess.Remove(nextElement)

		log.Debugf("Looking for posts in '%v'", nextDir)

		// If directory contains a post add to the list and stop exploring it
		post, err := newPost(nextDir)
		if err == nil {
			posts = append(posts, post)
			log.Debugf("Found post in '%v'", nextDir)
			continue
		}

		// Recursively check sub-direcotories for posts
		fileList, err := ioutil.ReadDir(nextDir)
		if err != nil {
			log.Warnf("Could not list files in directory '%v':"+err.Error(), nextDir)
			continue
		}
		for _, file := range fileList {
			if file.IsDir() {
				newDir := filepath.Join(nextDir, file.Name())
				toProcess.PushBack(newDir)
			}
		}
	}

	log.Infof("Found %v posts in '%v'", len(posts), baseDir)
	return posts
}

func newPost(dir string) (*Post, error) {
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// If the directory contains a metadata file it's a post directory
	var metadataFile string
	for _, file := range fileList {
		if metaDataMatch.MatchString(file.Name()) {
			metadataFile = filepath.Join(dir, file.Name())
		}
	}

	if metadataFile == "" {
		return nil, fmt.Errorf("Dir '%v' is not a post directory", dir)
	}

	return &Post{dir}, nil
}
