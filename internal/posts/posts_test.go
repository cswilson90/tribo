package posts

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPosts(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	posts := findPosts("testdata")
	assert.Equal(t, 3, len(posts), "Incorrect number of posts found")
}
