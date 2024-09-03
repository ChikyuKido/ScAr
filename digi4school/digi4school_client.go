package digi4school

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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
		"Connection":       "keep-alive",
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
		"Connection":      "keep-alive",
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
	_, _ = c.lti2Request(oauthMap2)

	return BookCookies{}, fmt.Errorf("failed to retrieve book cookie")
}

func (c *Digi4SchoolClient) lti1Request(params map[string]string) (map[string]string, error) {
	baseUrl := "https://kat.digi4school.at/lti"
	var queryParams []string
	queryParams = append(queryParams, "resource_link_id="+url.QueryEscape(params["resource_link_id"]))
	queryParams = append(queryParams, "lti_version="+url.QueryEscape(params["lti_version"]))
	queryParams = append(queryParams, "lti_message_type="+url.QueryEscape(params["lti_message_type"]))
	queryParams = append(queryParams, "user_id="+url.QueryEscape(params["user_id"]))
	queryParams = append(queryParams, "oauth_callback="+url.QueryEscape(params["oauth_callback"]))
	queryParams = append(queryParams, "oauth_nonce="+url.QueryEscape(params["oauth_nonce"]))
	queryParams = append(queryParams, "oauth_version="+url.QueryEscape(params["oauth_version"]))
	queryParams = append(queryParams, "oauth_timestamp="+url.QueryEscape(params["oauth_timestamp"]))
	queryParams = append(queryParams, "oauth_consumer_key="+url.QueryEscape(params["oauth_consumer_key"]))
	queryParams = append(queryParams, "oauth_signature_method="+url.QueryEscape(params["oauth_signature_method"]))
	queryParams = append(queryParams, "context_id="+url.QueryEscape(params["context_id"]))
	queryParams = append(queryParams, "context_type="+url.QueryEscape(params["context_type"]))
	queryParams = append(queryParams, "roles="+url.QueryEscape(params["roles"]))
	queryParams = append(queryParams, "oauth_signature="+url.QueryEscape(params["oauth_signature"]))

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
	req.Header.Set("Connection", "keep-alive")
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
	queryParams = append(queryParams, "resource_link_id="+url.QueryEscape(params["resource_link_id"]))
	queryParams = append(queryParams, "lti_version="+url.QueryEscape(params["lti_version"]))
	queryParams = append(queryParams, "lti_message_type="+url.QueryEscape(params["lti_message_type"]))
	queryParams = append(queryParams, "user_id="+url.QueryEscape(params["user_id"]))
	queryParams = append(queryParams, "oauth_callback="+url.QueryEscape(params["oauth_callback"]))
	queryParams = append(queryParams, "oauth_nonce="+url.QueryEscape(params["oauth_nonce"]))
	queryParams = append(queryParams, "oauth_version="+url.QueryEscape(params["oauth_version"]))
	queryParams = append(queryParams, "oauth_timestamp="+url.QueryEscape(params["oauth_timestamp"]))
	queryParams = append(queryParams, "oauth_consumer_key="+url.QueryEscape(params["oauth_consumer_key"]))
	queryParams = append(queryParams, "oauth_signature_method="+url.QueryEscape(params["oauth_signature_method"]))
	queryParams = append(queryParams, "context_id="+url.QueryEscape(params["context_id"]))
	queryParams = append(queryParams, "context_type="+url.QueryEscape(params["context_type"]))
	queryParams = append(queryParams, "roles="+url.QueryEscape(params["roles"]))
	queryParams = append(queryParams, "custom_code="+url.QueryEscape(params["custom_code"]))
	queryParams = append(queryParams, "custom_download="+url.QueryEscape(params["custom_download"]))
	queryParams = append(queryParams, "custom_warn="+url.QueryEscape(params["custom_warn"]))
	queryParams = append(queryParams, "oauth_signature="+url.QueryEscape(params["oauth_signature"]))

	encodedFormData := strings.Join(queryParams, "&")
	fmt.Println(encodedFormData)
	req, err := http.NewRequest("POST", baseUrl, bytes.NewBufferString(encodedFormData))
	if err != nil {
		log.Fatal(err)
	}

	cc, _ := url.Parse("https://a.digi4school.at")
	digi4s := ""
	for _, cookie := range c.Client.Jar.Cookies(cc) {
		if cookie.Name == "digi4s" {
			digi4s = cookie.Value
		}
	}
	fmt.Println(digi4s)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; rv:129.0) Gecko/20100101 Firefox/129.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/png,image/svg+xml,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://kat.digi4school.at")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Cookie", fmt.Sprintf("digi4s=%s", digi4s))

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(req.Header)
	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	fmt.Println(len(body))
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		fmt.Println(cookie.Name + ":" + cookie.Value)
	}

	return BookCookies{}, nil
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
	req.Header.Set("Connection", "keep-alive")
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
