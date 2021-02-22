package config

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const defaultConfigFile = ".tribo.yaml"

type TriboConfig struct {
	BlogName string `yaml:"blogTitle"`

	OutputDir   string `yaml:"outputDir"`
	PostsDir    string `yaml:"postsDir"`
	StaticDir   string `yaml:"staticDir"`
	TemplateDir string `yaml:"templateDir"`
}

var (
	// Values contains all the config variables for Tribo.
	Values TriboConfig

	// defaultConfig defines the default values for the config.
	defaultConfig = TriboConfig{
		BlogName: "My Blog",

		OutputDir:   "blog",
		PostsDir:    "posts",
		StaticDir:   "static",
		TemplateDir: "templates",
	}
)

// Init reads config values from the config file and command line and populates Values.
// The config is loaded from default values then the config file then the command line
// with duplicates being overwritten by the last place to specify them.
// The command line arguments (minus the program name) should be given as an argument to
// this function.
func Init(cmdArgs []string) {
	log.SetLevel(log.DebugLevel)

	Values = defaultConfig

	// Set up and parse flags
	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	configFile := flags.String("configFile", defaultConfigFile, "config file")

	outputDir := flags.String("outputDir", "", "output directory")
	postsDir := flags.String("postsDir", "", "posts directory")
	staticDir := flags.String("staticDir", "", "static files directory")
	templateDir := flags.String("templateDir", "", "template directory")
	flags.Parse(cmdArgs)

	// Load values from config file into Values
	loadConfigFile(*configFile)

	// Overwrite values in Values with those from command line if they've been given
	if *outputDir != "" {
		Values.OutputDir = *outputDir
	}
	if *postsDir != "" {
		Values.PostsDir = *postsDir
	}
	if *staticDir != "" {
		Values.StaticDir = *staticDir
	}
	if *templateDir != "" {
		Values.TemplateDir = *templateDir
	}

	// Convert file/path arguments into absolute paths
	Values.OutputDir = absPath(Values.OutputDir)
	Values.PostsDir = absPath(Values.PostsDir)
	Values.StaticDir = absPath(Values.StaticDir)
	Values.TemplateDir = absPath(Values.TemplateDir)
}

func absPath(file string) string {
	absPath, err := filepath.Abs(file)
	if err != nil {
		log.Fatalf("Failed to get absolute path of '%v'", file)
	}

	return absPath
}

func loadConfigFile(configFile string) {
	configFileGiven := configFile != defaultConfigFile

	absConfig, err := filepath.Abs(configFile)
	if err != nil {
		log.Fatalf("Failed to get absolute path of '%v'", configFile)
	}

	// Check if config file exists
	if _, err := os.Stat(absConfig); os.IsNotExist(err) {
		if configFileGiven {
			log.Fatalf("Config file '%v' doesn't exist", absConfig)
		}
		return
	}

	configYAML, err := ioutil.ReadFile(absConfig)
	if err != nil {
		log.Errorf(err.Error())
		return
	}

	err = yaml.Unmarshal(configYAML, &Values)
	if err != nil {
		log.Errorf("Error parsing YAML file '%v': "+err.Error(), absConfig)
	}
}
