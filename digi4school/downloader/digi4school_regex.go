package downloader

import (
	"regexp"
)

func CheckForEmbeddedImages(bodyString string) [][]string {
	pattern := `xlink:href="([^"]+\.(jpg|png))"`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(bodyString, -1)
	return matches
}

func GetDirName(url string) string {
	// regex to
	pattern := `\d+/(img|shade)/`

	// Compile the regex pattern
	re := regexp.MustCompile(pattern)

	// Find the first match
	match := re.FindString(url)

	return match
}
