package posts

import (
	"bufio"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gomarkdown/markdown"
	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/config"
)

type postData struct {
	PageTitle   string
	PostTitle   string
	PostContent template.HTML
}

type postListData struct {
	Title string
	Url   string
}

// postToHTML generates a posts HTML content and writes it to an output file.
func postToHTML(post *Post, outputFilename string) error {
	// Read markdown content and convert to HTML
	mdContent, err := ioutil.ReadFile(post.contentFile)
	if err != nil {
		return err
	}
	postHTML := markdown.ToHTML(mdContent, nil, nil)

	tmplData := postData{
		PageTitle:   post.metadata.title,
		PostTitle:   post.metadata.title,
		PostContent: template.HTML(postHTML),
	}

	return renderTemplate("post.html.tmpl", outputFilename, tmplData)
}

func postListHTML(posts Posts, outputFilename string) error {
	tmplData := make([]postListData, len(posts))
	for i, post := range posts {
		tmplData[i] = postListData{
			Title: post.metadata.title,
			Url:   post.urlPath,
		}
	}

	return renderTemplate("post_list.html.tmpl", outputFilename, tmplData)
}

func renderTemplate(templateName, outputFilename string, tmplData interface{}) error {
	templateFile := filepath.Join(config.Values.TemplateDir, templateName)
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	log.Debugf("Outputing template to: '%v'", outputFilename)
	outputWriter := bufio.NewWriter(outputFile)
	defer outputWriter.Flush()

	err = tmpl.Execute(outputWriter, tmplData)
	if err != nil {
		return err
	}

	return nil
}
