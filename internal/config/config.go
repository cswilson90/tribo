package config

import (
	"flag"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var (
	OutputDir   = "blog"
	PostsDir    = "posts"
	TemplateDir = "templates"
)

func Init() {
	log.SetLevel(log.DebugLevel)

	OutputDir = parseDir("outputDir", OutputDir, "output directory")
	PostsDir = parseDir("postsDir", PostsDir, "posts directory")
	TemplateDir = parseDir("templateDir", TemplateDir, "template directory")
}

func parseDir(name, defaultValue, description string) string {
	dir := flag.String(name, defaultValue, description)

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Failed to get absolute path of '%v'", dir)
	}

	return absDir
}
