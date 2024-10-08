package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

var Cookies []*http.Cookie

func downloadEmbeddedAsset(url string, matches [][]string) {
	trimmedURL := url[:strings.LastIndex(url, "/")+1]
	for _, match := range matches {
		if len(match) > 1 {
			downloadFile(fmt.Sprintf(trimmedURL + match[1]))
		}
	}
}

// function used to download one asset file (ex. embedded images)
func downloadFile(url string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panicln("failed to create request")
	}

	// add all cookies saved in cookies array (generated by CreateCookie function)
	for _, cookie := range Cookies {
		req.AddCookie(cookie)
	}

	// execute request and save response
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln("failed to get file:", err)
	}
	defer resp.Body.Close()

	dirname := GetDirName(url)

	// create file
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.MkdirAll(dirname, 0700) // Create your file
	}

	file, err := os.Create(dirname + path.Base(url))
	if err != nil {
		log.Panicln("failed to create file")
	}
	defer file.Close()

	// write contens of response to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Panicln("failed to copy file content:", err)
	}

}

func DownloadOnePage(url string) (string, error) {
	// create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panicln("failed to create request")
		return "", err
	}

	// set all cookies created from createCookie function
	for _, cookie := range Cookies {
		req.AddCookie(cookie)
	}

	client := &http.Client{}

	// execute request and save response
	resp, err := client.Do(req)
	if err != nil {
		log.Panicln("failed to get file:", err)
		return "", err
	}
	defer resp.Body.Close()

	// check if page exists
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("ERR 404 - %s", url)
	}

	// Convert the body to a string
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error
		log.Panicln("Error:", err)
	}

	// Convert the body to a string
	bodyString := string(bodyBytes)

	//check if page contains embedded mages
	if strings.Contains(bodyString, "image") {
		matches := CheckForEmbeddedImages(bodyString)
		if len(matches) > 0 {
			downloadEmbeddedAsset(url, matches)
		}
	}
	// set filename
	filename := path.Base(url)
	var parts = strings.Split(filename, ".")
	var length = len(parts[0])
	var number = strings.Repeat("0", 5-length) + parts[0]
	filename = number + "." + parts[1]

	// create file
	file, err := os.Create(filename)
	if err != nil {
		log.Panicln("failed to create file")
		return "", err
	}
	defer file.Close()

	// write contens of response to file
	_, err = io.WriteString(file, bodyString)
	if err != nil {
		log.Panicln("failed to copy file content:", err)
		return "", err
	}
	return filename, nil
}
