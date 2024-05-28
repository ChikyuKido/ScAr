package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

// CoursesResponse represents the response from the Moodle endpoint.
// It includes either a list of courses or an error message.
type CoursesResponse struct {
	Courses []Course `json:"courses"`
	Error   string   `json:"error"`
}

// Course represents a course the user is enrolled in.
type Course struct {
	ID              int             `json:"id"`
	Fullname        string          `json:"fullname"`
	ShortName       string          `json:"shortname"`
	Summary         string          `json:"summary"`
	Visible         bool            `json:"visible"`
	CourseImage     string          `json:"courseimage"`
	CourseImageType string          `json:"courseimagetype"`
	Category        string          `json:"coursecategory"`
	Sections        []CourseSection `json:"sections"`
}

// CourseSection represents a section within a course.
// Each section contains multiple modules.
type CourseSection struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	SectionNumber int            `json:"section"`
	Modules       []CourseModule `json:"modules"`
}

// CourseModule represents a module within a section.
// It includes details like the module name, description, URL, and its contents.
type CourseModule struct {
	ID          int                `json:"id"`
	Description string             `json:"description"`
	URL         string             `json:"url"`
	Name        string             `json:"name"`
	ModIcon     string             `json:"modicon"`
	ModName     string             `json:"modname"`
	Dates       []CourseModuleDate `json:"dates"`
	Contents    []CourseContent    `json:"contents"`
}

// CourseContent represents the content of a module.
// It includes information such as the type, filename, file size, and URL.
type CourseContent struct {
	Type     string `json:"type"`
	Filename string `json:"filename"`
	FileSize int64  `json:"size"`
	FileURL  string `json:"fileurl"`
}

// CourseModuleDate represents a date associated with a course module.
// It includes a label, timestamp, and data ID.
type CourseModuleDate struct {
	Label     string `json:"label"`
	Timestamp int64  `json:"timestamp"`
	DataID    string `json:"dataid"`
}

type CourseApi struct {
	client *MoodleClient
}

func newCourseApi(client *MoodleClient) *CourseApi {
	return &CourseApi{client}
}

func (courseApi *CourseApi) GetCourses(fetchSectionsInCourse bool) ([]Course, error) {
	body, err := courseApi.client.makeWebserviceRequest("core_course_get_enrolled_courses_by_timeline_classification", map[string]string{"classification": "all"})

	if err != nil {
		return nil, err
	}
	var coursesResp CoursesResponse
	if err := json.Unmarshal([]byte(body), &coursesResp); err != nil {
		return nil, err
	}

	logrus.Info("Found ", len(coursesResp.Courses), " Course")
	for i := range coursesResp.Courses {
		if strings.Contains(coursesResp.Courses[i].CourseImage, "data:image/svg+xml;base64,") {
			coursesResp.Courses[i].CourseImageType = "BASE64"
		} else {
			coursesResp.Courses[i].CourseImageType = "URL"
		}
	}

	if fetchSectionsInCourse {
		logrus.Info("Fetch sections for Courses")
		total := len(coursesResp.Courses)
		for i := range coursesResp.Courses {
			progress := (i + 1) * 100 / total
			fmt.Printf("\rFetch Section for Course [%d|%d]: [%-50s] %d%%", i+1, total, strings.Repeat("=", progress/2), progress)
			err := courseApi.FetchCourseContents(&coursesResp.Courses[i])
			if err != nil {
				logrus.Error("Could not get section data for course ", coursesResp.Courses[i].Fullname, "(", coursesResp.Courses[i].ID, ")", ": ", err)
			}
		}
		fmt.Println()

	}
	return coursesResp.Courses, nil
}

// FetchCourseContents This method fetches the contents for a course and adds it automatically to the course.
// This method can be automatically called in the getAllCourses
func (courseApi *CourseApi) FetchCourseContents(course *Course) error {
	var body, err = courseApi.client.makeWebserviceRequest("core_course_get_contents", map[string]string{"courseid": strconv.Itoa(course.ID)})

	if err != nil {
		return err
	}
	//When the json starts with a bracket then it's an error because a valid start with a [
	if body[0] == '{' {
		return fmt.Errorf("error in response json: %v", string(body))
	}
	var sections []CourseSection
	if err := json.Unmarshal(body, &sections); err != nil {
		return err
	}
	course.Sections = sections

	return nil
}

func (courseApi *CourseApi) GetAssignModule(module CourseModule) error {
	var body, err = courseApi.client.makeWebserviceRequest("mod_assign_get_assignments", map[string]string{"courseids[0]": strconv.Itoa(module.ID)})

	if err != nil {
		return err
	}

	logrus.Info(string(body))
	return nil

}
