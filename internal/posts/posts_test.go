package posts

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const inputDir = "testdata"

// TODO when output is produced
// * Add test for duplicate checking

func TestPosts(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	posts := findPosts(inputDir)
	assert.Equal(t, 3, len(posts), "Incorrect number of posts found")

	tmpDir := t.TempDir()
	BuildPosts(inputDir, tmpDir)
}
