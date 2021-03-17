package posts

import (
	"bufio"
	"encoding/xml"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/config"
)

const RSSDateFormat = "Mon, 02 Jan 2006 15:04:05 MST"

type RSSXML struct {
	XMLName xml.Name    `xml:"rss"`
	Version string      `xml:"version,attr"`
	Channel *ChannelXML `xml:"channel"`
}

type ChannelXML struct {
	XMLName       xml.Name   `xml:"channel"`
	Title         string     `xml:"title"`
	Link          string     `xml:"link"`
	Description   string     `xml:"description"`
	LastBuildDate string     `xml:"lastBuildDate"`
	PubDate       string     `xml:"pubDate"`
	TTL           int        `xml:"ttl"`
	Items         []*ItemXML `xml:"item"`
}

type ItemXML struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Guid        string   `xml:"guid"`
	PubDate     string   `xml:pubDate`
}

// postRSSFeed outputs the RSS feed for the blog.
// posts should be presorted by date
func postRSSFeed(posts Posts, outputFile string) {
	log.Infof("Writing RSS XML to '%v'", outputFile)

	// Add newest 10 posts to RSS feed
	maxPosts := 10
	if len(posts) < maxPosts {
		maxPosts = len(posts)
	}

	postsXML := make([]*ItemXML, maxPosts)
	for i := 0; i < maxPosts; i++ {
		post := posts[i]
		postsXML[i] = &ItemXML{
			Title:       post.metadata.title,
			Link:        post.urlPath,
			Description: "",
			Guid:        post.urlPath,
			PubDate:     post.metadata.publishDate.Format(RSSDateFormat),
		}
	}

	lastDate := time.Now()
	if maxPosts > 0 {
		lastDate = posts[0].metadata.publishDate
	}

	channelXML := &ChannelXML{
		Title:         config.Values.BlogName,
		Link:          config.Values.BaseUrlPath,
		Description:   config.Values.BlogName,
		LastBuildDate: lastDate.Format(RSSDateFormat),
		PubDate:       time.Now().Format(RSSDateFormat),
		TTL:           1800,
		Items:         postsXML,
	}

	rssXML := &RSSXML{
		Version: "2.0",
		Channel: channelXML,
	}

	xmlFile, err := os.Create(outputFile)
	if err != nil {
		log.Errorf("Failed to open RSS file '%v': "+err.Error(), outputFile)
		return
	}
	defer xmlFile.Close()

	xmlWriter := bufio.NewWriter(xmlFile)
	xmlWriter.WriteString(xml.Header)
	xmlEncoder := xml.NewEncoder(xmlWriter)
	xmlEncoder.Indent("", "  ")

	// Encode calls Flush on Writer
	err = xmlEncoder.Encode(rssXML)
	if err != nil {
		log.Errorf("Failed to write RSS file: " + err.Error())
	}
}
