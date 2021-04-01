package posts

import (
	"bufio"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/config"
)

// commonData contains template data comon to rendering all the post templates.
type commonData struct {
	// BaseUrlPath is the path prefix used when serving the blog.
	// e.g. if blog is at http://example.com/blog/ the BaseUrlPath would be "/blog"
	BaseUrlPath     string
	BlogName        string
	BlogDescription string
	CurrentYear     string
	// PageTitle is the HTML title of the page.
	PageTitle string
}

// postData contains the template data for a single blog post.
type postData struct {
	Title       string
	Content     template.HTML
	Preview     template.HTML
	PublishDate string
	// Url is the URL used to link to the post.
	Url  string
	Tags []string
}

// postListPageData contains all the template data for rendering the post list page.
type postListPageData struct {
	Common commonData
	Posts  []postData
	// AllTags is a list of unique tags from all the posts that are in the post list.
	AllTags []string
}

// postPageData contains all the template data for rendering a single blog post page.
type postPageData struct {
	Common commonData
	Post   postData
}

// tmpl stores the parsed templates used to render all post output.
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
// It uses the "post.html.tmpl" template.
func postToHTML(post *Post, outputFilename string) error {
	postData := postToPostData(post, false)
	tmplData := postPageData{
		Common: comData(),
		Post:   postData,
	}

	tmplData.Common.PageTitle = post.title

	return renderTemplate("post.html.tmpl", outputFilename, tmplData)
}

// postListHTML generates the HTML for the list of posts used as the main page for the blog.
// It uses the "post_list.html.tmpl" template file.
func postListHTML(posts Posts, outputFilename string) error {
	tmplData := postListPageData{
		Common: comData(),
		Posts:  make([]postData, len(posts)),
	}

	uniqueTags := make(map[string]struct{})
	for i, post := range posts {
		tmplData.Posts[i] = postToPostData(post, true)

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

// renderTemplate renders a template and saves the output to a file.
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

// postToPostData generates a postData object from a post.
func postToPostData(post *Post, previewContent bool) postData {
	return postData{
		Title:       post.title,
		Content:     template.HTML(post.content),
		Preview:     template.HTML(post.preview),
		PublishDate: post.metadata.publishDate.Format("2 Jan 2006"),
		Url:         post.urlPath,
		Tags:        post.metadata.tags,
	}
}

// commonData returns a commonData object that can be used when rendering a template.
func comData() commonData {
	return commonData{
		BaseUrlPath:     config.Values.BaseUrlPath,
		BlogName:        config.Values.BlogName,
		BlogDescription: config.Values.BlogDescription,
		CurrentYear:     time.Now().Format("2006"),
		PageTitle:       config.Values.BlogName,
	}
}
