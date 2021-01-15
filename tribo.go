package tribo

import (
	"path/filepath"
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/cswilson90/tribo/internal/posts"
)

func RunTribo() {
	postsDir := flag.String("postsDir", "posts", "posts directory")
	absPath, err := filepath.Abs(filepath.FromSlash(*postsDir))
	if err != nil {
		log.Fatalf("Invalid posts directory given:"+err.Error())
	}
	posts := posts.FindPosts(absPath)

	log.Infof("Found %v posts", len(posts))
}
