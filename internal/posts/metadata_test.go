package posts

import (
	"fmt"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var today = time.Now().Format(dateFormat)

var tests = []struct {
	dir   string
	title string
	date  string
	tags  []string
}{
	{"testdata/2021/01/post1/", "2021 01 Post 1", "2021-01-24", []string{"happy", "upbeat"}},
	{"testdata/2021/01/post2/", "2021 01 Post 2", today, []string{"jolly"}},
	{"testdata/2020/12/post2/", "2020 12 Post 2", "2020-12-04", nil},
}

var errorTests = []struct {
	dir string
}{
	{"testdata/2020/12/not-post/"},
	{"testdata/errors/no-title/"},
	{"testdata/errors/invalid-yaml/"},
	{"testdata/errors/invalid-json/"},
	{"testdata/errors/invalid-date/"},
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
