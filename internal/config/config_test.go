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
			OutputDir:   "blog",
			PostsDir:    "posts",
			StaticDir:   "static",
			TemplateDir: "templates",
			Parallelism: runtime.NumCPU(),
		},
	},
	{
		flags: []string{
			"-outputDir", "/home/test/output",
			"-postsDir", "other/posts",
			"-parallelism", "8",
		},
		expectedValues: TriboConfig{
			BlogName:    "My Blog",
			OutputDir:   "/home/test/output",
			PostsDir:    "other/posts",
			StaticDir:   "static",
			TemplateDir: "templates",
			Parallelism: 8,
		},
	},
	{
		flags: []string{"-configFile", "testdata/test_config.yaml", "-outputDir", "/home/test/output"},
		expectedValues: TriboConfig{
			BlogName:    "My Blog",
			OutputDir:   "/home/test/output",
			PostsDir:    "posts",
			StaticDir:   "static",
			TemplateDir: "other/templates",
			Parallelism: runtime.NumCPU(),
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
