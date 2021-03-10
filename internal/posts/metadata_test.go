package posts

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	dir      string
	title    string
	linkName string
	date     string
	tags     []string
}{
	{
		dir:      "testdata/posts/2021/01/post1/",
		title:    "2021 01 Post 1",
		linkName: "2021-01-post-1",
		date:     "2021-01-24",
		tags:     []string{"happy", "upbeat"},
	},
	{
		dir:      "testdata/posts/2021/01/post2/",
		title:    "2021 01 Post 2",
		linkName: "post2-2021-01",
		date:     "2021-01-01",
		tags:     []string{"jolly"},
	},
	{
		dir:      "testdata/posts/2020/12/post2/",
		title:    "2020 12 Post 2",
		linkName: "post-2-202012",
		date:     "2020-12-04",
		tags:     nil,
	},
}

var errorTests = []struct {
	dir string
}{
	{"testdata/posts/2020/12/not-post/"},
	{"testdata/posts/errors/no-title/"},
	{"testdata/posts/errors/invalid-yaml/"},
	{"testdata/posts/errors/invalid-json/"},
	{"testdata/posts/errors/invalid-date/"},
}

func TestMetadata(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	assert := assert.New(t)

	for i, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("Test %v", i), func(t *testing.T) {
			t.Parallel()

			metaData, err := parseMetadata(tc.dir)
			if err != nil {
				t.Errorf("Couldn't load metadata: " + err.Error())
			}

			assert.Equal(tc.title, metaData.title, "Title incorrect")
			assert.Equal(tc.linkName, metaData.linkName, "Link name incorrect")
			assert.Equal(tc.date, metaData.publishDate.Format(dateFormat), "Date incorrect")
			assert.Equal(tc.tags, metaData.tags, "Tags incorrect")
		})
	}

	for i, tc := range errorTests {
		tc := tc
		t.Run(fmt.Sprintf("Test %v", i), func(t *testing.T) {
			t.Parallel()

			_, err := parseMetadata(tc.dir)
			if err == nil {
				t.Errorf("Expected error test %v", i)
			}
		})
	}
}
