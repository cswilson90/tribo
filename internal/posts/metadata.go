package posts

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	dateFormat        = "2006-01-02"
	linkNameMaxLength = 50
)

var (
	// Dangerous characters that can cause problems in file names and URLs
	linkNameDangerous = regexp.MustCompile(`[/?.:=%#\t\n]`)
	metadataMatch     = regexp.MustCompile(`^metadata\.(json|ya?ml)$`)
)

// PostMetadata stores the metadata about a post.
type PostMetadata struct {
	linkName    string
	publishDate time.Time
	tags        []string
	title       string
}

// rawPostMetaData defines the structure of metadata in the config file.
type rawPostMetadata struct {
	LinkName    string
	PublishDate string
	Tags        []string
	Title       string
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
	if rawData.Title == "" {
		return nil, fmt.Errorf("No title given for post")
	}

	// Default publish time to today if not given
	publishTime := time.Now()

	var err error
	if rawData.PublishDate != "" {
		publishTime, err = time.Parse(dateFormat, rawData.PublishDate)
		if err != nil {
			return nil, fmt.Errorf("Could not parse publish date '%v': "+err.Error(), rawData.PublishDate)
		}
	}

	// If no link name has been given make one from the title
	if rawData.LinkName == "" {
		linkRunes := []rune(linkNameDangerous.ReplaceAllString(rawData.Title, ""))
		maxLength := linkNameMaxLength
		if len(linkRunes) < maxLength {
			maxLength = len(linkRunes)
		}
		rawData.LinkName = string(linkRunes[:maxLength])
	}

	// Remove potentially dangerous characters from link name, convert spaces
	// to dashes and lowercase
	rawData.LinkName = linkNameDangerous.ReplaceAllString(rawData.LinkName, "")
	rawData.LinkName = strings.ToLower(strings.ReplaceAll(rawData.LinkName, " ", "-"))

	// Sort tags
	sort.Strings(rawData.Tags)

	return &PostMetadata{
		linkName:    rawData.LinkName,
		title:       rawData.Title,
		publishDate: publishTime,
		tags:        rawData.Tags,
	}, nil
}
