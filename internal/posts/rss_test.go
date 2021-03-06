package posts

import (
	"bufio"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/cswilson90/tribo/internal/config"
)

const (
	blogName        = "Test Blog"
	blogDescription = "My test blog"
	baseUrlPath     = "/blog"
	rssLinkUrl      = "https://test.invalid"
)

func TestRSS(t *testing.T) {
	config.Values.BlogName = blogName
	config.Values.BlogDescription = blogDescription
	config.Values.BaseUrlPath = baseUrlPath
	config.Values.RssLinkUrl = rssLinkUrl

	posts := Posts{
		&Post{
			urlPath: "/2021/03/test-post-1",
			metadata: &PostMetadata{
				publishDate: time.Date(2021, time.March, 17, 0, 0, 0, 0, time.UTC),
			},
			title:     "Test Post 1",
			preview:   "<p>Preview Paragraph</p>",
			published: true,
		},
		&Post{
			urlPath: "/2021/02/test-post-2",
			metadata: &PostMetadata{
				publishDate: time.Date(2021, time.February, 24, 0, 0, 0, 0, time.UTC),
			},
			title:     "Test Post 2",
			preview:   " <p> Description  </p>",
			published: true,
		},
		&Post{
			urlPath: "/2021/01/test-post-3",
			metadata: &PostMetadata{
				publishDate: time.Date(2021, time.January, 12, 0, 0, 0, 0, time.UTC),
			},
			title:     "Test Post 3",
			preview:   "<p> Description Paragraph</p> ",
			published: true,
		},
	}

	expectedDescriptions := []string{
		"Preview Paragraph",
		"Description",
		"Description Paragraph",
	}

	tmpDir := t.TempDir()
	rssFileName := filepath.Join(tmpDir, "rss.xml")
	postRSSFeed(posts, rssFileName)

	rssFile, err := os.Open(rssFileName)
	if err != nil {
		t.Fatalf("Failed to read RSS file '%v': %v", rssFileName, err.Error())
	}

	rssFileReader := bufio.NewReader(rssFile)
	rssDecoder := xml.NewDecoder(rssFileReader)
	rssXML := &RSSXML{}

	err = rssDecoder.Decode(rssXML)
	if err != nil {
		t.Fatalf("Failed to parse RSS file '%v': %v", rssFileName, err.Error())
	}

	assert := assert.New(t)
	assert.Equal("2.0", rssXML.Version, "Incorrect RSS version")

	channel := rssXML.Channel
	assert.Equal(blogName, channel.Title, "Incorrect RSS channel title")
	assert.Equal(rssLinkUrl+baseUrlPath, channel.Link, "Incorrect RSS channel Link")
	assert.Equal(blogDescription, channel.Description, "Incorrect RSS channel description")
	assert.Equal(posts[0].metadata.publishDate.Format(RSSDateFormat), channel.LastBuildDate, "Incorrect RSS channel build date")
	assert.Equal(1800, channel.TTL, "Incorrect RSS channel TTL")

	items := channel.Items
	if len(items) != len(posts) {
		t.Fatalf("Expected %v items in RSS XML got %v", len(posts), len(items))
	}

	for i, item := range items {
		assert.Equal(posts[i].title, item.Title, "Incorrect title for post %v", i)
		assert.Equal(rssLinkUrl+posts[i].urlPath, item.Link, "Incorrect link for post %v", i)
		assert.Equal(expectedDescriptions[i], item.Description, "Incorrect description for post %v", i)
		assert.Equal(rssLinkUrl+posts[i].urlPath, item.Guid, "Incorrect guid for post %v", i)
		assert.Equal(posts[i].metadata.publishDate.Format(RSSDateFormat), item.PubDate, "Incorrect pubdate for post %v", i)
	}
}
