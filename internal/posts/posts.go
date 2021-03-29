package posts

import (
	"container/list"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/config"
)

type (
	DirSet     map[string]bool
	Posts      []*Post
	renderMode int
)

const (
	linkNameMaxLength = 50

	// Render the full post except the title
	renderPost renderMode = 0
	// Render a preview of the post (the first paragraph)
	renderPreview renderMode = 1
	// Render the title of the post (the first heading)
	renderTitle renderMode = 2
)

// Keeps track of seen directories to catch duplicates
var (
	uniqueDirs     = make(map[string]*Post)
	uniqueDirsLock sync.Mutex

	// Dangerous characters that can cause problems in file names and URLs
	linkNameDangerous = regexp.MustCompile(`[/?.:=%#\t\n]`)

	// Regexes for checking if string looks like year or month
	looksLikeYear  = regexp.MustCompile(`^\d{4}$`)
	looksLikeMonth = regexp.MustCompile(`^\d{2}$`)
)

type Post struct {
	dir         string
	outputDir   string
	contentFile string
	resourceDir string

	urlPath  string
	metadata *PostMetadata

	content  string
	preview  string
	title    string
	linkName string

	published bool
}

// postRenderer renders a post in HTML, supports rendering previews.
// Implements markdown.Rendere
type postRenderer struct {
	htmlRenderer *html.Renderer

	mode      renderMode
	rendering bool

	seenTitle bool
}

// Functions to sort a list of posts by publish date with newest first
func (p Posts) Len() int      { return len(p) }
func (p Posts) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Posts) Less(i, j int) bool {
	return p[i].metadata.publishDate.After(p[j].metadata.publishDate)
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
	numWorkers := config.Values.Parallelism
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

	// Filter out unpublished posts
	publishedPosts := make(Posts, 0)
	for _, post := range posts {
		if post.published {
			publishedPosts = append(publishedPosts, post)
		}
	}

	sort.Sort(publishedPosts)

	// Remove directories from output that don't have a published post
	err = removeExtraOutputDirs(outputDir)
	if err != nil {
		log.Errorf("Failed to clean up non-existent posts from output directory: %v", err.Error())
	}

	// Output list of posts HTML
	indexFile := filepath.Join(absOutputDir, "index.html")
	postListHTML(publishedPosts, indexFile)

	// Output RSS feed of posts
	rssFile := filepath.Join(absOutputDir, "rss.xml")
	postRSSFeed(publishedPosts, rssFile)
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
	resourceDir := ""
	for _, file := range fileList {
		if isMetadataFile(file.Name()) {
			metadata = true
		}
		if file.Name() == "content.md" {
			content = filepath.Join(dir, file.Name())
		}

		if file.Name() == "resources" && file.IsDir() {
			resourceDir = filepath.Join(dir, file.Name())
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
		resourceDir: resourceDir,
	}, nil
}

// build builds the post from the directory set in the object.
func (p *Post) build(outputDir string) error {
	var err error
	p.metadata, err = parseMetadata(p.dir)
	if err != nil {
		return err
	}

	// Check if post should be published yet
	// Publish all posts of futurePosts config option has been set
	if !config.Values.FuturePosts && p.metadata.publishDate.After(time.Now()) {
		return nil
	}

	// Parse markdown content and save
	mdContent, err := ioutil.ReadFile(p.contentFile)
	if err != nil {
		return err
	}
	p.content = parsePostMarkdown(mdContent, renderPost)
	p.preview = parsePostMarkdown(mdContent, renderPreview)
	p.title = parsePostMarkdown(mdContent, renderTitle)

	// If no link name has been given make one from the title
	p.linkName = p.metadata.linkName
	if p.linkName == "" {
		linkRunes := []rune(linkNameDangerous.ReplaceAllString(p.title, ""))
		maxLength := linkNameMaxLength
		if len(linkRunes) < maxLength {
			maxLength = len(linkRunes)
		}
		p.linkName = string(linkRunes[:maxLength])
	}

	// Remove potentially dangerous characters from link name, convert spaces
	// to dashes and lowercase
	p.linkName = linkNameDangerous.ReplaceAllString(p.linkName, "")
	p.linkName = strings.ToLower(strings.ReplaceAll(p.linkName, " ", "-"))

	// Build filepath for post from publish date and linkname
	year := p.metadata.publishDate.Format("2006")
	month := p.metadata.publishDate.Format("01")
	p.urlPath = strings.Join([]string{config.Values.BaseUrlPath, year, month, p.linkName}, "/")
	p.outputDir = filepath.Join(outputDir, year, month, p.linkName)

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

	// Copy static resources to output file if the post has some
	if p.resourceDir != "" {
		log.Debugf("Copying post resources from '%v'", p.resourceDir)
		err = copy.Copy(p.resourceDir, p.outputDir)
		if err != nil {
			log.Errorf("Failed to copy resource files from '%v' to '%v':"+err.Error(), p.resourceDir, p.outputDir)
		}
	}

	// Generate HTML of post from markdown and templates
	indexFile := filepath.Join(p.outputDir, "index.html")
	err = postToHTML(p, indexFile)
	if err != nil {
		return err
	}

	p.published = true
	return nil
}

func parsePostMarkdown(mdContent []byte, mode renderMode) string {
	opts := html.RendererOptions{Flags: html.CommonFlags}
	renderer := &postRenderer{
		htmlRenderer: html.NewRenderer(opts),
		mode:         mode,
	}

	return string(markdown.ToHTML(mdContent, nil, renderer))
}

// removeExtraOutputDirs removes directories from the output that don't have a post
// from the current run.
func removeExtraOutputDirs(outputDir string) error {
	outputFileList, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return err
	}

	uniqueDirsLock.Lock()
	defer uniqueDirsLock.Unlock()

	for _, outputFile := range outputFileList {
		if outputFile.IsDir() && looksLikeYear.MatchString(outputFile.Name()) {
			yearDir := filepath.Join(outputDir, outputFile.Name())
			yearFileList, err := ioutil.ReadDir(yearDir)
			if err != nil {
				return err
			}

			for _, yearFile := range yearFileList {
				if yearFile.IsDir() && looksLikeMonth.MatchString(yearFile.Name()) {
					monthDir := filepath.Join(yearDir, yearFile.Name())
					monthFileList, err := ioutil.ReadDir(monthDir)
					if err != nil {
						return err
					}

					for _, monthFile := range monthFileList {
						if monthFile.IsDir() {
							postDir := filepath.Join(monthDir, monthFile.Name())
							if uniqueDirs[postDir] == nil {
								os.RemoveAll(postDir)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// markdown.Renderer.RenderNode() implementation
func (r *postRenderer) RenderNode(w io.Writer, node ast.Node, entering bool) ast.WalkStatus {
	if r.mode == renderPost {
		r.rendering = true
		// Render whole post except title (the first heading)
		if !r.seenTitle {
			switch node.(type) {
			case *ast.Heading:
				if entering {
					return ast.SkipChildren
				} else {
					r.seenTitle = true
					return ast.GoToNext
				}
			}
		}
	} else if r.mode == renderPreview {
		// Render a preview of the post (the first paragraph)
		switch node.(type) {
		case *ast.Paragraph:
			if entering {
				// Entering first paragraph so start rendering
				r.rendering = true
			} else {
				// Leaving first paragraph so render end and terminate
				r.htmlRenderer.RenderNode(w, node, entering)
				r.rendering = false
				return ast.Terminate
			}
		}
	} else if r.mode == renderTitle {
		// Render the title of the post (the first heading)
		switch node.(type) {
		case *ast.Heading:
			if entering {
				// Entering first heading so start rendering but don't add opening tag
				r.rendering = true
				return ast.GoToNext
			} else {
				// Leaving first heading so we've rendered it all
				r.rendering = false
				return ast.Terminate
			}
		}
	} else {
		log.Fatalf("Unknown render mode %v", r.mode)
	}

	if r.rendering {
		return r.htmlRenderer.RenderNode(w, node, entering)
	} else {
		return ast.GoToNext
	}
}

// markdown.Renderer.RenderHeader() implementation
func (r *postRenderer) RenderHeader(w io.Writer, ast ast.Node) {
	r.htmlRenderer.RenderHeader(w, ast)
}

// markdown.Renderer.RenderFooter() implementatio
func (r *postRenderer) RenderFooter(w io.Writer, ast ast.Node) {
	r.htmlRenderer.RenderFooter(w, ast)
}
