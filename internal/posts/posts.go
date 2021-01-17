package posts

import (
	"container/list"
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

func FindPosts(baseDir string) Posts {
	log.Infof("Looking for posts recursivly in '%v'", baseDir)

	// Save which directories we've seen before to avoid circular references
	// causing an infinite loop
	seenLists := make(DirSet)

	toProcess := list.New()
	toProcess.PushBack(baseDir)
	seenLists[baseDir] = true

	posts := make(Posts, 0)

	for toProcess.Len() > 0 {
		nextElement := toProcess.Front()
		nextDir := nextElement.Value.(string)
		toProcess.Remove(nextElement)

		log.Debugf("Looking for posts in '%v'", nextDir)

		fileList, err := ioutil.ReadDir(nextDir)
		if err != nil {
			log.Warnf("Could not list files in directory '%v':"+err.Error(), nextDir)
			continue
		}

		// If the directory contains a metadata file it's a post directory
		for _, file := range fileList {
			if metaDataMatch.MatchString(file.Name()) {
				posts = append(posts, &Post{nextDir})
				continue
			}
		}

		// Recursively check sub-direcotories for posts
		for _, file := range fileList {
			if file.IsDir() {
				newDir := filepath.Join(nextDir, file.Name())
				if !seenLists[newDir] {
					toProcess.PushBack(newDir)
					seenLists[newDir] = true
				}
			}
		}

		log.Debugf("Found %v posts in '%v'", len(posts), nextDir)
	}

	log.Infof("Found %v posts in '%v'", len(posts), baseDir)
	return posts
}
