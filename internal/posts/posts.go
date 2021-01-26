package posts

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"
)

type DirSet map[string]bool

type Posts []*Post

type Post struct {
	dir      string
	metadata *PostMetadata
}

// BuildPosts finds all the posts in a directory and builds them.
// The provided directory is converted to an absolute directory before use.
func BuildPosts(baseDir string) {
	absDir, err := filepath.Abs(baseDir)
	if err != nil {
		log.Fatalf("Failed to absolute path of dir '%v': "+err.Error(), baseDir)
	}

	posts := findPosts(absDir)

	// Build posts in parallel
	// TODO make number of workers configurable
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	postJobs := make(chan *Post, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go buildPosts(postJobs, &wg)
	}

	for _, post := range posts {
		postJobs <- post
	}
	close(postJobs)
	wg.Wait()
}

func buildPosts(posts <-chan *Post, wg *sync.WaitGroup) {
	for {
		post, more := <-posts
		if !more {
			return
		}
		post.build()
	}
}

// FindPosts recursively searches a directory for posts.
func findPosts(baseDir string) Posts {
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

// newPost checks if the given directory is a post directory and creates a post if so.
// Returns an error if the directory is not a post directory.
func newPost(dir string) (*Post, error) {
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// If the directory contains a metadata file and a content file it's a post directory
	metadata, content := false, false
	for _, file := range fileList {
		if isMetadataFile(file.Name()) {
			metadata = true
		}
		if file.Name() == "content.md" {
			content = true
		}
	}

	if !metadata && !content {
		return nil, fmt.Errorf("Dir '%v' is not a post directory", dir)
	}

	if !(metadata && content) {
		log.Errorf("Dir '%v' is missing metadata or content file", dir)
		return nil, fmt.Errorf("Dir '%v' is not a post directory", dir)
	}

	return &Post{
		dir: dir,
	}, nil
}

// build builds the post from the directory set in the object.
func (p *Post) build() {
	var err error
	p.metadata, err = parseMetadata(p.dir)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
}
