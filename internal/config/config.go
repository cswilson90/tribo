/*
	Package config parses the Tribo config file and command line arguments and
	provides access to the values.

	The Init() function should be called before trying to use any of the config
	values.

	All values should be accessed from the Values variable.

		import (
			"os"

			"github.com/cswilson90/tribo/internal/config"
		)

		config.Init(os.Args[1:])
		baseURLPath := config.Values.BaseUrlPath
*/
package config

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const defaultConfigFile = ".tribo.yaml"

// TriboConfig stores all config values for Tribo.
type TriboConfig struct {
	/*
		BaseUrlPath is the base path of the blog on the hostname where it's hosted.
		For example if the blog is hosted at http://example.com/blog/ you would set
		this config value to "/blog".
		The value can be left blank if the blog is served from the root of the site.
	*/
	BaseUrlPath     string `yaml:"baseUrlPath"`
	BlogName        string `yaml:"blogName"`
	BlogDescription string `yaml:"blogDescription"`

	// NoRss controls whether an RSS feed is generated.
	// The default value is false so an RSS feed will be generated.
	NoRss bool `yaml:"noRss"`
	/*
		RssLinkURL is used to build absolute links when linking to posts in the RSS feed.
		It should be set to the scheme and hostname of the URL the blog is served from.
		The BaseURLPath is appended to this value to build the links.
		For example if the blog is served from http://example.com/blog/ you should set
		this value to "http://example.com/".
		Can be ignored if NoRss is set to true.
	*/
	RssLinkUrl string `yaml:"rssLinkUrl"`

	OutputDir   string `yaml:"outputDir"`
	PostsDir    string `yaml:"postsDir"`
	StaticDir   string `yaml:"staticDir"`
	TemplateDir string `yaml:"templateDir"`

	// Parallelism controls the max number of blog posts built in parallel.
	// Defaults to the number of CPUs available on the machine.
	Parallelism int `yaml:"parallelism"`
	// FuturePosts controls whether blog posts with a publish date set in the future
	// are published.
	FuturePosts bool `yaml:"futurePosts"`
	// NoOutputCleanup controls whether Tribo tries to clean up old blog posts in the output.
	// By default Tribo will delete any directories from the output directory that it thinks are
	// from posts which no longer exist or have been moved due to a title or published date change.
	// You can set this option to true to stop this behaviour if it is causing problems.
	NoOutputCleanup bool `yaml:"noOutputCleanup"`
}

var (
	/*
		Values contains all the config variables for Tribo.
		Other parts of the program should access the config values using this variable.

			import "github.com/cswilson/tribo/internal/config"

			baseURLPath := config.Values.BaseUrlPath

	*/
	Values TriboConfig

	// defaultConfig defines the default values for the config.
	defaultConfig = TriboConfig{
		BlogName:        "My Blog",
		BlogDescription: "My musings about the world",

		RssLinkUrl: "http://127.0.0.1",

		OutputDir:   "blog",
		PostsDir:    "posts",
		StaticDir:   "static",
		TemplateDir: "templates",

		Parallelism:     runtime.NumCPU(),
		FuturePosts:     false,
		NoOutputCleanup: false,
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

	blogName := flags.String("blogName", "", "blog name")
	blogDescription := flags.String("blogDescription", "", "blog description")
	baseUrlPath := flags.String("baseUrlPath", "", "base blog URL path")

	noRss := flags.Bool("noRss", false, "don't generate an RSS feed")
	rssLinkUrl := flags.String("rssLinkUrl", "", "RSS link base URL")

	outputDir := flags.String("outputDir", "", "output directory")
	postsDir := flags.String("postsDir", "", "posts directory")
	staticDir := flags.String("staticDir", "", "static files directory")
	templateDir := flags.String("templateDir", "", "template directory")

	parallelism := flags.Int("parallelism", 0, "max parallelism")
	futurePosts := flags.Bool("futurePosts", false, "publish future posts")
	noOutputCleanup := flags.Bool("noOutputCleanup", false, "don't attempt to clean up output directory")
	flags.Parse(cmdArgs)

	// Load values from config file into Values
	loadConfigFile(*configFile)

	// Overwrite values in Values with those from command line if they've been given
	if *blogName != "" {
		Values.BlogName = *blogName
	}
	if *blogDescription != "" {
		Values.BlogDescription = *blogDescription
	}
	if *baseUrlPath != "" {
		Values.BaseUrlPath = *baseUrlPath
	}
	if *noRss {
		Values.NoRss = *noRss
	}
	if *rssLinkUrl != "" {
		Values.RssLinkUrl = *rssLinkUrl
	}
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
	if *parallelism != 0 {
		Values.Parallelism = *parallelism
	}
	if *futurePosts {
		Values.FuturePosts = *futurePosts
	}
	if *noOutputCleanup {
		Values.NoOutputCleanup = *noOutputCleanup
	}

	// Convert file/path arguments into absolute paths
	Values.OutputDir = absPath(Values.OutputDir)
	Values.PostsDir = absPath(Values.PostsDir)
	Values.StaticDir = absPath(Values.StaticDir)
	Values.TemplateDir = absPath(Values.TemplateDir)
}

// absPath converts a file path to an absolute path.
// If the file path cannot be converted then the program will exit with an error.
func absPath(file string) string {
	absPath, err := filepath.Abs(file)
	if err != nil {
		log.Fatalf("Failed to get absolute path of '%v'", file)
	}

	return absPath
}

// loadConfigFile loads the config from a file into the Values variable.
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
