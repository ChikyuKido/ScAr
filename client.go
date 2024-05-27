package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

type TokenResponse struct {
	Token        string `json:"token"`
	PrivateToken string `json:"privatetoken"`
	Error        string `json:"error"`
	ErrorCode    string `json:"errorcode"`
}

type MoodleClient struct {
	ServiceUrl   string
	Token        string
	PrivateToken string
	Username     string
	SkipSSL      bool
	CourseApi    *CourseApi
}

func NewMoodleClient(serviceUrl string, skipSSL bool) *MoodleClient {
	if skipSSL {
		logrus.Info("Skipping SSL verification for all requests")
	}
	client := &MoodleClient{ServiceUrl: serviceUrl, SkipSSL: skipSSL}
	client.CourseApi = newCourseApi(client)
	return client
}

func (mc *MoodleClient) login(username string, password string) error {
	loginURL := fmt.Sprintf("%s/login/token.php", mc.ServiceUrl)

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("service", "moodle_mobile_app")
	req, err := http.NewRequest("POST", loginURL, nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = data.Encode()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: mc.SkipSSL},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}

	if tokenResp.Error != "" {
		return fmt.Errorf("failed to obtain token: %s", tokenResp.Error)
	}

	logrus.Info("Successfully logged in")

	mc.Token = tokenResp.Token
	mc.PrivateToken = tokenResp.PrivateToken
	mc.Username = username
	return nil
}
