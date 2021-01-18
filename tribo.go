package tribo

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/posts"
)

func RunTribo() {
	postsDir := flag.String("postsDir", "posts", "posts directory")
	posts.BuildPosts(*postsDir)
}
