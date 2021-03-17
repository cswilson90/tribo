package config

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

var tests = []struct {
	flags          []string
	expectedValues TriboConfig
}{
	{
		flags: []string{},
		expectedValues: TriboConfig{
			BlogName:    "My Blog",
			RssLinkUrl:  "http://127.0.0.1",
			OutputDir:   "blog",
			PostsDir:    "posts",
			StaticDir:   "static",
			TemplateDir: "templates",
			Parallelism: runtime.NumCPU(),
			FuturePosts: false,
		},
	},
	{
		flags: []string{
			"-outputDir", "/home/test/output",
			"-postsDir", "other/posts",
			"-parallelism", "8",
			"-futurePosts",
			"-rssLinkUrl", "https://example.com",
		},
		expectedValues: TriboConfig{
			BlogName:    "My Blog",
			RssLinkUrl:  "https://example.com",
			OutputDir:   "/home/test/output",
			PostsDir:    "other/posts",
			StaticDir:   "static",
			TemplateDir: "templates",
			Parallelism: 8,
			FuturePosts: true,
		},
	},
	{
		flags: []string{"-configFile", "testdata/test_config.yaml", "-outputDir", "/home/test/output"},
		expectedValues: TriboConfig{
			BlogName:    "My Blog",
			RssLinkUrl:  "http://127.0.0.1",
			OutputDir:   "/home/test/output",
			PostsDir:    "posts",
			StaticDir:   "static",
			TemplateDir: "other/templates",
			Parallelism: runtime.NumCPU(),
			FuturePosts: true,
		},
	},
}

func TestConfig(t *testing.T) {
	assert := assert.New(t)

	for i, tc := range tests {
		Init(tc.flags)

		expected := tc.expectedValues
		expected.OutputDir = absPath(expected.OutputDir)
		expected.PostsDir = absPath(expected.PostsDir)
		expected.StaticDir = absPath(expected.StaticDir)
		expected.TemplateDir = absPath(expected.TemplateDir)

		assert.Equal(expected, Values, fmt.Sprintf("Test %v unexpected result", i+1))
	}
}
