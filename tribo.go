package tribo

import (
	"github.com/cswilson90/tribo/internal/config"
	"github.com/cswilson90/tribo/internal/posts"
)

func RunTribo() {
	config.Init()
	posts.BuildPosts(config.Values.PostsDir, config.Values.OutputDir)
}
