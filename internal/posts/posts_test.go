package posts

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const inputDir = "testdata"

// TODO when output is produced
// * Add test for duplicate checking

func TestFindPosts(t *testing.T) {
	log.SetLevel(log.FatalLevel)

	posts := findPosts(inputDir)
	assert.Equal(t, 3, len(posts), "Incorrect number of posts found")
}

func TestBuildPosts(t *testing.T) {
	log.SetLevel(log.FatalLevel)

	tmpDir := t.TempDir()
	BuildPosts(inputDir, tmpDir)
}
