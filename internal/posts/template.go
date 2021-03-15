package posts

import (
	"bufio"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/gomarkdown/markdown"
	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/config"
)

type commonData struct {
	BaseUrlPath string
	BlogName    string
	PageTitle   string
}

type postData struct {
	Title       string
	Content     template.HTML
	PublishDate string
	Url         string
	Tags        []string
}

// Passed to post_list.html.tmpl
type postListPageData struct {
	Common  commonData
	Posts   []postData
	AllTags []string
}

// Passed to post.html.tmpl
type postPageData struct {
	Common commonData
	Post   postData
}

var tmpl *template.Template

// initTemplates initialises the Template variable for use when generating posts.
// This function needs to be called before generating post output files.
func initTemplates() error {
	includesPattern := filepath.Join(config.Values.TemplateDir, "includes", "*.html.tmpl")
	var err error
	tmpl, err = template.ParseGlob(includesPattern)
	if err != nil {
		return err
	}

	templatePattern := filepath.Join(config.Values.TemplateDir, "*.html.tmpl")
	tmpl, err = tmpl.ParseGlob(templatePattern)
	if err != nil {
		return err
	}

	return nil
}

// postToHTML generates a posts HTML content and writes it to an output file.
func postToHTML(post *Post, outputFilename string) error {
	postData, err := postToPostData(post)
	if err != nil {
		return err
	}
	tmplData := postPageData{
		Common: comData(),
		Post:   postData,
	}

	tmplData.Common.PageTitle = post.metadata.title

	return renderTemplate("post.html.tmpl", outputFilename, tmplData)
}

// postListHTML generates the HTML for the list of posts used as the main page for the blog.
func postListHTML(posts Posts, outputFilename string) error {
	tmplData := postListPageData{
		Common: comData(),
		Posts:  make([]postData, len(posts)),
	}

	uniqueTags := make(map[string]struct{})
	for i, post := range posts {
		var err error
		tmplData.Posts[i], err = postToPostData(post)
		if err != nil {
			return err
		}

		for _, tag := range tmplData.Posts[i].Tags {
			uniqueTags[tag] = struct{}{}
		}
	}

	tmplData.AllTags = make([]string, len(uniqueTags))
	i := 0
	for tag := range uniqueTags {
		tmplData.AllTags[i] = tag
		i++
	}
	sort.Strings(tmplData.AllTags)

	return renderTemplate("post_list.html.tmpl", outputFilename, tmplData)
}

func renderTemplate(templateName, outputFilename string, tmplData interface{}) error {
	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	log.Debugf("Outputing template to: '%v'", outputFilename)
	outputWriter := bufio.NewWriter(outputFile)
	defer outputWriter.Flush()

	err = tmpl.ExecuteTemplate(outputWriter, templateName, tmplData)
	if err != nil {
		return err
	}

	return nil
}

func postToPostData(post *Post) (postData, error) {
	// Read markdown content and convert to HTML
	mdContent, err := ioutil.ReadFile(post.contentFile)
	if err != nil {
		return postData{}, err
	}
	postHTML := markdown.ToHTML(mdContent, nil, nil)

	return postData{
		Title:       post.metadata.title,
		Content:     template.HTML(postHTML),
		PublishDate: post.metadata.publishDate.Format("2 Jan 2006"),
		Url:         post.urlPath,
		Tags:        post.metadata.tags,
	}, nil
}

func comData() commonData {
	return commonData{
		BaseUrlPath: config.Values.BaseUrlPath,
		BlogName:    config.Values.BlogName,
		PageTitle:   config.Values.BlogName,
	}
}
