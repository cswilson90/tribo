package tribo

import (
	"os"

	"github.com/cswilson90/tribo/internal/config"
	"github.com/cswilson90/tribo/internal/posts"
)

func RunTribo() {
	config.Init(os.Args[1:])
	posts.BuildPosts(config.Values.PostsDir, config.Values.OutputDir)
}
