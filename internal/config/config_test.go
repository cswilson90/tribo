package config

import (
	"fmt"
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
			OutputDir:   "blog",
			PostsDir:    "posts",
			TemplateDir: "templates",
		},
	},
	{
		flags: []string{"-outputDir", "/home/test/output", "-postsDir", "other/posts"},
		expectedValues: TriboConfig{
			OutputDir:   "/home/test/output",
			PostsDir:    "other/posts",
			TemplateDir: "templates",
		},
	},
	{
		flags: []string{"-configFile", "testdata/test_config.yaml", "-outputDir", "/home/test/output"},
		expectedValues: TriboConfig{
			OutputDir:   "/home/test/output",
			PostsDir:    "posts",
			TemplateDir: "other/templates",
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
		expected.TemplateDir = absPath(expected.TemplateDir)

		assert.Equal(expected, Values, fmt.Sprintf("Test %v unexpected result", i+1))
	}
}
