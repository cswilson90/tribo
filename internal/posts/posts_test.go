package posts

import (
	"os"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/cswilson90/tribo/internal/config"
)

const inputDir = "testdata/posts"
const staticDir = "testdata/static"
const templateDir = "testdata/templates"

func TestFindPosts(t *testing.T) {
	log.SetLevel(log.FatalLevel)

	posts := findPosts(inputDir)
	assert.Equal(t, 3, len(posts), "Incorrect number of posts found")
}

func TestBuildPosts(t *testing.T) {
	config.Init([]string{})
	log.SetLevel(log.FatalLevel)
	config.Values.StaticDir = staticDir
	config.Values.TemplateDir = templateDir

	tmpDir := t.TempDir()
	BuildPosts(inputDir, tmpDir)

	expectedDirs := []string{
		"2021/01/2021-01-post-1/",
		"2021/01/post2-2021-01/",
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

	expectedRootFiles := []string{"index.html", "test.css", "rss.xml"}
	for _, file := range expectedRootFiles {
		mainIndex := filepath.Join(tmpDir, file)
		if _, err := os.Stat(mainIndex); os.IsNotExist(err) {
			t.Errorf("Expected html file '%v' doesn't exist", mainIndex)
		}
	}

	expectedResourceFiles := []string{"2020/12/post-2-202012/static.file"}
	for _, file := range expectedResourceFiles {
		resourceFile := filepath.Join(tmpDir, file)
		if _, err := os.Stat(resourceFile); os.IsNotExist(err) {
			t.Errorf("Expected resource file '%v' doesn't exist", resourceFile)
		}
	}
}
