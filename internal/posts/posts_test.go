package posts

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPosts(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	posts := FindPosts("testdata")
	assert.Equal(t, 3, len(posts), "Incorrect number of posts found")
}
