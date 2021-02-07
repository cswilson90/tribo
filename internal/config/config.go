package config

import (
	"flag"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type TriboConfig struct {
	OutputDir   string
	PostsDir    string
	TemplateDir string
}

// Values contains all the config variables for Tribo.
var Values = TriboConfig{
	OutputDir:   "blog",
	PostsDir:    "posts",
	TemplateDir: "templates",
}

func Init() {
	log.SetLevel(log.DebugLevel)

	Values.OutputDir = parseDir("outputDir", Values.OutputDir, "output directory")
	Values.PostsDir = parseDir("postsDir", Values.PostsDir, "posts directory")
	Values.TemplateDir = parseDir("templateDir", Values.TemplateDir, "template directory")
}

func parseDir(name, defaultValue, description string) string {
	dir := flag.String(name, defaultValue, description)

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Failed to get absolute path of '%v'", dir)
	}

	return absDir
}
