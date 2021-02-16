package posts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/cswilson90/tribo/internal/config"
)

const inputDir = "testdata/posts"
const templateDir = "testdata/templates"

// TODO when output is produced
// * Add test for duplicate checking

func TestFindPosts(t *testing.T) {
	log.SetLevel(log.FatalLevel)

	posts := findPosts(inputDir)
	assert.Equal(t, 3, len(posts), "Incorrect number of posts found")
}

func TestBuildPosts(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	config.Values.TemplateDir = templateDir

	tmpDir := t.TempDir()
	BuildPosts(inputDir, tmpDir)

	today := time.Now()
	todayDir := filepath.Join(today.Format("2006"), today.Format("01"))

	expectedDirs := []string{
		"2021/01/2021-01-post-1/",
		filepath.Join(todayDir, "post2-2021-01/"),
		"2020/12/post-2-202012/",
	}

	for _, dir := range expectedDirs {
		joinedDir := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(joinedDir); os.IsNotExist(err) {
			t.Errorf("Expected directory '%v' doesn't exist", joinedDir)
		}

		indexFile := filepath.Join(joinedDir, "index.html")
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			t.Errorf("Expected html file '%v' doesn't exist", indexFile)
		}
	}

	mainIndex := filepath.Join(tmpDir, "index.html")
	if _, err := os.Stat(mainIndex); os.IsNotExist(err) {
		t.Errorf("Expected html file '%v' doesn't exist", mainIndex)
	}
}
