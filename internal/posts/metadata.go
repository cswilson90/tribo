package posts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// dateFormat specifies the expected format of publishDates in the metadata.
const dateFormat = "2006-01-02"

var metadataMatch = regexp.MustCompile(`^metadata\.(json|ya?ml)$`)

// PostMetadata stores the metadata about a post.
type PostMetadata struct {
	linkName    string
	publishDate time.Time
	tags        []string
}

// rawPostMetaData defines the structure of metadata in the config file.
type rawPostMetadata struct {
	LinkName    string
	PublishDate string
	Tags        []string
}

// isMetadataFile returns true if a file is a metadata file.
func isMetadataFile(filename string) bool {
	return metadataMatch.MatchString(filename)
}

// parseMetadata parses the metadata file from a directory.
func parseMetadata(dir string) (*PostMetadata, error) {
	fileList, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Look for metadata files, there should be exactly one
	metaFiles := make([]string, 0)
	for _, file := range fileList {
		if isMetadataFile(file.Name()) {
			metaFiles = append(metaFiles, file.Name())
		}
	}

	if len(metaFiles) > 1 {
		return nil, fmt.Errorf("Found multiple metadata files in '%v'", dir)
	} else if len(metaFiles) == 0 {
		return nil, fmt.Errorf("No metadata files found in '%v'", dir)
	}

	metaFile := metaFiles[0]
	fileExt := filepath.Ext(metaFile)

	fullPath := filepath.Join(dir, metaFile)
	data, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	rawMetadata := &rawPostMetadata{}
	if fileExt == ".json" {
		err = json.Unmarshal(data, rawMetadata)
	} else if fileExt == ".yaml" || fileExt == ".yml" {
		err = yaml.Unmarshal(data, rawMetadata)
	} else {
		log.Fatalf("Got unknown metadata file extension '%v'", fileExt)
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to parse metadata '%v': "+err.Error(), fullPath)
	}

	metadata, err := processRawMetadata(rawMetadata)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse metadata '%v': "+err.Error(), fullPath)
	}

	return metadata, nil
}

// processRawMetadata converts the raw data to the right types and does validation.
func processRawMetadata(rawData *rawPostMetadata) (*PostMetadata, error) {
	if rawData.PublishDate == "" {
		return nil, fmt.Errorf("No publish date given for post")
	}

	publishTime, err := time.Parse(dateFormat, rawData.PublishDate)
	if err != nil {
		return nil, fmt.Errorf("Could not parse publish date '%v': "+err.Error(), rawData.PublishDate)
	}

	// Sort tags
	sort.Strings(rawData.Tags)

	return &PostMetadata{
		linkName:    rawData.LinkName,
		publishDate: publishTime,
		tags:        rawData.Tags,
	}, nil
}
