package digi4school

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"scar/digi4school/downloader"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Digi4SchoolClient struct {
	Username string
	Password string
	Client   *http.Client
}

type BookCookies struct {
	Digi4Bname  string
	Digi4Bvalue string
	Digi4Pname  string
	Digi4Pvalue string
	Path        string
	SubPath     string
}

type Book struct {
	Name     string
	DataCode string
	DataId   string
}

func NewDigi4SClient(username, password string) *Digi4SchoolClient {
	transport := http.DefaultTransport
	jar, _ := cookiejar.New(nil)
	return &Digi4SchoolClient{
		Username: username,
		Password: password,
		Client: &http.Client{
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: transport,
		},
	}
}

func (c *Digi4SchoolClient) Login() error {
	baseUrl := "https://digi4school.at/br/xhr/login"

	payload := url.Values{}
	payload.Set("email", c.Username)
	payload.Set("password", c.Password)

	headers := map[string]string{
		"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0",
		"Accept":           "text/plain, */*; q=0.01",
		"Accept-Language":  "en-US,en;q=0.5",
		"Referer":          "https://digi4school.at/",
		"Content-Type":     "application/x-www-form-urlencoded; charset=UTF-8",
		"X-Requested-With": "XMLHttpRequest",
		"Origin":           "https://digi4school.at",
	}

	req, err := http.NewRequest("POST", baseUrl, strings.NewReader(payload.Encode()))
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for _, cookie := range c.Client.Jar.Cookies(req.URL) {
		if cookie.Name == "digi4s" {
			return nil
		}
	}
	return fmt.Errorf("login failed")
}

func (c *Digi4SchoolClient) Logout() error {
	baseUrl := "https://digi4school.at/br/logout"

	headers := map[string]string{
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0",
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.5",
		"Referer":         "https://digi4school.at/",
	}

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Digi4SchoolClient) GetBooks() ([]Book, error) {
	baseUrl := "https://digi4school.at/ebooks"
	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Priority", "u=0, i")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		panic(err)
	}

	books := make([]Book, 0)
	div := doc.Find("#shelf")
	div.Find("a.bag").Each(func(index int, item *goquery.Selection) {
		dataCode, _ := item.Attr("data-code")
		dataID, _ := item.Attr("data-id")
		bookName := item.Find("h1").Text()
		books = append(books, Book{
			Name:     bookName,
			DataCode: dataID,
			DataId:   dataCode,
		})
	})
	return books, nil
}

func (c *Digi4SchoolClient) DownloadBook(book Book) error {
	fmt.Println("Downloading book: " + book.Name)
	bookCookies, _ := c.getBookCookie(book.DataId)

	// create temp dir
	tmp, err := os.MkdirTemp(os.TempDir(), "bookdl_*")
	if err != nil {
		log.Panicln(err)
	}
	// save current dir
	current, err := os.Getwd()
	if err != nil {
		log.Panicln(err)
	}
	// change into temp dir
	os.Chdir(tmp)
	if err != nil {
		log.Panicln(err)
	}

	digi4bCookie := &http.Cookie{Name: bookCookies.Digi4Bname, Value: bookCookies.Digi4Bvalue}
	digi4pCookie := &http.Cookie{Name: bookCookies.Digi4Pname, Value: bookCookies.Digi4Pvalue}

	downloader.Cookies = make([]*http.Cookie, 0)
	digi4sCookie := &http.Cookie{Name: "digi4s", Value: c.getCurrentDigi4sCookie()}
	fmt.Println(digi4sCookie)
	downloader.Cookies = append(downloader.Cookies, digi4pCookie, digi4bCookie, digi4sCookie)
	defer os.Chdir(current)
	defer os.RemoveAll(tmp)
	page := 1
	for {
		// https://a.digi4school.at/ebook/7010/1/index.html
		var baseUrl = ""
		if bookCookies.SubPath != "" {
			baseUrl = fmt.Sprintf("https://a.digi4school.at/ebook/%s/%s", book.DataCode, bookCookies.SubPath)
		} else {
			baseUrl = fmt.Sprintf("https://a.digi4school.at/ebook/%s", book.DataCode)
		}
		err := downloader.DownloadOnePage(fmt.Sprintf("%s/%d.svg", baseUrl, page))
		if err != nil {
			fmt.Println(err)
			break
		}
		page++
	}
	return nil
}

func (c *Digi4SchoolClient) getBookCookie(buchId string) (BookCookies, error) {

	oauthMap, err := c.getOauthMap(buchId)
	if err != nil {
		return BookCookies{}, fmt.Errorf("could not refresh digi4s cookie: %v", err)
	}

	oauthMap2, _ := c.lti1Request(oauthMap)
	finishedCookies, _ := c.lti2Request(oauthMap2)

	return finishedCookies, fmt.Errorf("failed to retrieve book cookie")
}

func (c *Digi4SchoolClient) lti1Request(params map[string]string) (map[string]string, error) {
	baseUrl := "https://kat.digi4school.at/lti"
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}

	encodedFormData := strings.Join(queryParams, "&")
	req, err := http.NewRequest("POST", baseUrl, bytes.NewBufferString(encodedFormData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://digi4school.at")
	req.Header.Set("Priority", "u=0, i")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return map[string]string{}, fmt.Errorf("could not load: %v", err)
	}
	return extractParams(string(body)), nil
}

func (c *Digi4SchoolClient) lti2Request(params map[string]string) (BookCookies, error) {
	baseUrl := "https://a.digi4school.at/lti"
	var queryParams []string
	for key, value := range params {
		queryParams = append(queryParams, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}

	encodedFormData := strings.Join(queryParams, "&")
	req, err := http.NewRequest("POST", baseUrl, bytes.NewBufferString(encodedFormData))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://kat.digi4school.at")
	req.Header.Set("Priority", "u=0, i")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	finishedCookies := BookCookies{}
	fmt.Println(resp.Status)

	finishedCookies.SubPath = c.checkSubPath(resp.Header.Get("Location"))

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "digi4b" {
			finishedCookies.Digi4Bvalue = cookie.Value
			finishedCookies.Digi4Bname = cookie.Name
			finishedCookies.Path = cookie.Path
		}
		if cookie.Name == "digi4p" {
			finishedCookies.Digi4Pvalue = cookie.Value
			finishedCookies.Digi4Pvalue = cookie.Name
		}
	}
	if finishedCookies.Digi4Bvalue == "" || finishedCookies.Digi4Pvalue == "" {
		//error handling
	}
	return finishedCookies, nil
}
func (c *Digi4SchoolClient) checkSubPath(url string) string {
	fmt.Println(url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://kat.digi4school.at")
	req.Header.Set("Priority", "u=0, i")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "sbnr") {
		return "1"
	}

	return ""
}

func (c *Digi4SchoolClient) getOauthMap(buchId string) (map[string]string, error) {
	baseUrl := "https://digi4school.at/ebook/" + buchId

	req, err := http.NewRequest("GET", baseUrl, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Priority", "u=0, i")

	resp, err := c.Client.Do(req)
	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return extractParams(string(body)), nil
}

func (c *Digi4SchoolClient) getCurrentDigi4sCookie() string {
	uri, _ := url.Parse("https://a.digi4school.at")
	for _, cookie := range c.Client.Jar.Cookies(uri) {
		if cookie.Name == "digi4s" {
			return cookie.Value
		}
	}
	return ""
}

func extractParams(formHTML string) map[string]string {
	params := make(map[string]string)
	re := regexp.MustCompile(`<input\s+name='([^']+)' value='([^']*)'>`)
	matches := re.FindAllStringSubmatch(formHTML, -1)
	for _, match := range matches {
		params[match[1]] = match[2]
	}
	return params
}
