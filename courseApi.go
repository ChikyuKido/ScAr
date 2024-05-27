package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type CoursesResponse struct {
	Courses []Course `json:"courses"`
}

type Course struct {
	ID              int    `json:"id"`
	Fullname        string `json:"fullname"`
	ShortName       string `json:"shortname"`
	Summary         string `json:"summary"`
	Visible         bool   `json:"visible"`
	CourseImage     string `json:"courseimage"`
	CourseImageType string `json:"courseimagetype"`
	Category        string `json:"coursecategory"`
}

type CourseApi struct {
	client *MoodleClient
}

func newCourseApi(client *MoodleClient) *CourseApi {
	return &CourseApi{client}
}

func (courseApi *CourseApi) getCourses() ([]Course, error) {
	webserviceURL := fmt.Sprintf("%s/webservice/rest/server.php", courseApi.client.ServiceUrl)

	data := url.Values{}
	data.Set("wstoken", courseApi.client.Token)
	data.Set("wsfunction", "core_course_get_enrolled_courses_by_timeline_classification") // TODO: make enum for other functions
	data.Set("moodlewsrestformat", "json")
	data.Set("classification", "all") // TODO: make enum for 'all', 'inprogress', 'future', 'past'

	req, err := http.NewRequest("GET", webserviceURL, nil)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = data.Encode()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: courseApi.client.SkipSSL},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var coursesResp CoursesResponse
	if err := json.Unmarshal(body, &coursesResp); err != nil {
		return nil, err
	}

	logrus.Info("Found ", len(coursesResp.Courses), " Course")
	for _, course := range coursesResp.Courses {
		if strings.Contains(course.CourseImage, "data:image/svg+xml;base64,") {
			course.CourseImageType = "BASE64"
		} else {
			course.CourseImageType = "URL"
		}
	}

	return coursesResp.Courses, nil
}
