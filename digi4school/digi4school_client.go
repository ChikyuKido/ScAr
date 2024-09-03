package digi4school

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Digi4SchoolClient struct {
	Username string
	Password string
	Client   *http.Client
}

type BookCookies struct {
	Digi4B string
	Digi4P string
}

func NewDigi4SClient(username, password string) *Digi4SchoolClient {
	transport := http.DefaultTransport
	envproxy := os.Getenv("HTTPS_PROXY")
	if envproxy != "" {
		proxyUrl, _ := url.Parse("https://" + envproxy)
		proxy := http.ProxyURL(proxyUrl)
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           proxy,
		} // for mitmprox
	}
	jar, _ := cookiejar.New(nil)
	return &Digi4SchoolClient{
		Username: username,
		Password: password,
		Client: &http.Client{
			Jar: jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}, // disable redirects
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

func (c *Digi4SchoolClient) GetBookCookie(buchId string) (BookCookies, error) {

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
	fmt.Println(encodedFormData)
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
	fmt.Println(encodedFormData)
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
	fmt.Println(resp.StatusCode, resp.Status)
	fmt.Println(resp.Header)
	//body, err := io.ReadAll(resp.Body)
	//fmt.Println(string(body))
	//io.Copy(os.Stdout, resp.Body)
	defer resp.Body.Close()
	fmt.Println("Cookies: ")

    finishedCookies := BookCookies{}
    for _, cookie := range resp.Cookies() {
		//fmt.Println(cookie.Name + ":" + cookie.Value)
	    if cookie.Name == "digi4b" {
            finishedCookies.Digi4B = cookie.Value
        }
        if cookie.Name == "digi4p"{
            finishedCookies.Digi4P = cookie.Value
        }
    }
    if finishedCookies.Digi4B == "" || finishedCookies.Digi4P == ""{
        //error handling
    }
	return finishedCookies, nil
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

func extractParams(formHTML string) map[string]string {
	params := make(map[string]string)
	re := regexp.MustCompile(`<input\s+name='([^']+)' value='([^']*)'>`)
	matches := re.FindAllStringSubmatch(formHTML, -1)
	for _, match := range matches {
		params[match[1]] = match[2]
	}
	return params
}


