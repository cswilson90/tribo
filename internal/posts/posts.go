package posts

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/config"
)

type (
	DirSet map[string]bool
	Posts  []*Post
)

// Keeps track of seen directories to catch duplicates
var uniqueDirs = make(map[string]*Post)
var uniqueDirsLock sync.Mutex

type Post struct {
	dir         string
	outputDir   string
	contentFile string

	urlPath  string
	metadata *PostMetadata
}

// BuildPosts finds all the posts in a directory and builds them.
// The provided directory is converted to an absolute directory before use.
func BuildPosts(inputDir, outputDir string) {
	absInputDir, err := filepath.Abs(inputDir)
	if err != nil {
		log.Fatalf("Failed to absolute path of dir '%v': "+err.Error(), inputDir)
	}
	absOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		log.Fatalf("Failed to absolute path of dir '%v': "+err.Error(), outputDir)
	}

	err = initTemplates()
	if err != nil {
		log.Fatalf("Failed to parse post templates: " + err.Error())
	}

	posts := findPosts(absInputDir)

	// Build posts in parallel
	// TODO make number of workers configurable
	numWorkers := runtime.NumCPU()
	if numWorkers > len(posts) {
		numWorkers = len(posts)
	}

	var wg sync.WaitGroup
	postJobs := make(chan *Post, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go buildPosts(absOutputDir, postJobs, &wg)
	}

	for _, post := range posts {
		postJobs <- post
	}
	close(postJobs)
	wg.Wait()

	// Copy static files to output dir
	log.Infof("Copying static files from '%v' to '%v'", config.Values.StaticDir, absOutputDir)
	err = copy.Copy(config.Values.StaticDir, absOutputDir)
	if err != nil {
		log.Fatalf("Failed to copy static files from '%v' to '%v':"+err.Error(), config.Values.StaticDir, absOutputDir)
	}

	// Output list of posts HTML
	indexFile := filepath.Join(absOutputDir, "index.html")
	postListHTML(posts, indexFile)
}

// buildPosts gets posts from a channel and builds them.
func buildPosts(outputDir string, posts <-chan *Post, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		post, more := <-posts
		if !more {
			return
		}
		err := post.build(outputDir)
		if err != nil {
			log.Errorf("Error building post in '%v': "+err.Error(), post.dir)
		}
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
	metadata, content := false, ""
	for _, file := range fileList {
		if isMetadataFile(file.Name()) {
			metadata = true
		}
		if file.Name() == "content.md" {
			content = filepath.Join(dir, file.Name())
		}
	}

	if !metadata && content == "" {
		return nil, fmt.Errorf("Dir '%v' is not a post directory", dir)
	}

	if !(metadata && content != "") {
		log.Errorf("Dir '%v' is missing metadata or content file", dir)
		return nil, fmt.Errorf("Dir '%v' is not a post directory", dir)
	}

	return &Post{
		dir:         dir,
		contentFile: content,
	}, nil
}

// build builds the post from the directory set in the object.
func (p *Post) build(outputDir string) error {
	var err error
	p.metadata, err = parseMetadata(p.dir)
	if err != nil {
		return err
	}

	// Build filepath for post from publish date and linkname
	year := p.metadata.publishDate.Format("2006")
	month := p.metadata.publishDate.Format("01")
	p.urlPath = strings.Join([]string{config.Values.BaseUrlPath, year, month, p.metadata.linkName}, "/")
	p.outputDir = filepath.Join(outputDir, year, month, p.metadata.linkName)

	// Do a uniqueness check on directory name
	uniqueDirsLock.Lock()
	duplicate, exists := uniqueDirs[p.outputDir]
	if exists {
		return fmt.Errorf("Same output directory as '%v'", duplicate.dir)
	}
	uniqueDirs[p.outputDir] = p
	uniqueDirsLock.Unlock()

	// Make output directory
	err = os.MkdirAll(p.outputDir, 0775)
	if err != nil {
		return err
	}

	// Generate HTML of post from markdown and templates
	indexFile := filepath.Join(p.outputDir, "index.html")
	err = postToHTML(p, indexFile)
	if err != nil {
		return err
	}

	return nil
}
