package tribo

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/posts"
)

func RunTribo() {
	postsDir := flag.String("postsDir", "posts", "posts directory")
	outputDir := flag.String("outputDir", "blog", "output directory")

	log.SetLevel(log.InfoLevel)

	posts.BuildPosts(*postsDir, *outputDir)
}
