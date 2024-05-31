package moodle

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"scar/util"
	"strconv"
	"strings"
)

// CoursesResponse represents the response from the Moodle endpoint.
// It includes either a list of courses or an error message.
type CoursesResponse struct {
	Courses []Course `json:"courses"`
	Error   string   `json:"error"`
}

type CoursesModAssignResponse struct {
	CourseModAssigns []CourseModAssign `json:"courses"`
	Error            string            `json:"error"`
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
	ComponentID int                `json:"id"`
	ID          int                `json:"instance"`
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
	client      *MoodleClient
	courseCache CourseCache
}

// CourseModAssign contains the data for assignments for one course
type CourseModAssign struct {
	ID          int                   `json:"id"`
	Assignments []CourseModAssignment `json:"assignments"`
}

type CourseModAssignment struct {
	ComponentID         int          `json:"cmid"`
	AssignmentID        int          `json:"id"`
	Intro               string       `json:"intro"`
	SubmissionStatement string       `json:"submissionstatement"`
	IntroAttachment     []MoodleFile `json:"introattachments"`
}

type MoodleFile struct {
	FileName string `json:"filename"`
	FileSize int64  `json:"filesize"`
	FileURL  string `json:"fileurl"`
}

type CourseCache struct {
	CourseModAssignments []CourseModAssignment `json:"assignments"`
}

type DownloadAssignmentData struct {
	ID                         int      `json:"id"`
	CMID                       int      `json:"cmid"`
	CourseID                   int      `json:"courseid"`
	Name                       string   `json:"name"`
	Intro                      string   `json:"intro"`
	ModName                    string   `json:"modname"`
	SubmissionStatement        string   `json:"submissionstatement"`
	IntroAttachmentsNames      []string `json:"introattachmentsnames"`
	SubmissionAttachmentsNames []string `json:"submissionattachmentsnames"`
}

func newCourseApi(client *MoodleClient) *CourseApi {
	return &CourseApi{client, CourseCache{}}
}

func (courseApi *CourseApi) GetCourses(fetchSectionsInCourse bool) ([]Course, error) {
	body, err := courseApi.client.makeWebserviceRequest("core_course_get_enrolled_courses_by_timeline_classification", map[string]string{"classification": "all"})

	if err != nil {
		return nil, err
	}
	var coursesResp CoursesResponse
	if err := json.Unmarshal(body, &coursesResp); err != nil {
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
	if len(courseApi.courseCache.CourseModAssignments) == 0 {
		logrus.Info("Assignments not cached yet. Requesting it now")
		var ids []string
		for _, course := range coursesResp.Courses {
			ids = append(ids, strconv.Itoa(course.ID))
		}
		var params = map[string]string{}
		for i := 0; i < len(ids); i++ {
			params[fmt.Sprintf("courseids[%d]", i)] = ids[i]
		}
		body, err = courseApi.client.makeWebserviceRequest("mod_assign_get_assignments", params)
		var courseModAssignResponse CoursesModAssignResponse
		if err := json.Unmarshal(body, &courseModAssignResponse); err != nil {
			return nil, err
		}
		if len(courseModAssignResponse.Error) != 0 {
			return nil, fmt.Errorf("%v", courseModAssignResponse.Error)
		}
		for _, course := range courseModAssignResponse.CourseModAssigns {
			courseApi.courseCache.CourseModAssignments = append(courseApi.courseCache.CourseModAssignments, course.Assignments...)
		}
		logrus.Info("Found ", len(courseApi.courseCache.CourseModAssignments), " Assignments")
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

func (courseApi *CourseApi) DownloadAssignModule(module *CourseModule, basePath string) error {
	if module.ModName != "assign" {
		logrus.Fatal("Module is not a assignment")
	}
	var body, err = courseApi.client.makeWebserviceRequest("mod_assign_get_submission_status", map[string]string{"assignid": strconv.Itoa(module.ID)})
	if err != nil {
		return err
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	var submissionMoodleFiles []MoodleFile
	if value, ok := result["lastattempt"]; ok {
		lastAttempt := value.(map[string]interface{})
		if value, ok := lastAttempt["submission"]; ok {
			submission := value.(map[string]interface{})
			if value, ok := submission["plugins"]; ok {
				plugins := value.([]interface{})

				for _, plugin := range plugins {
					pluginMap := plugin.(map[string]interface{})
					if pluginMap["type"] == "file" {
						fileAreas := pluginMap["fileareas"].([]interface{})
						for _, fileArea := range fileAreas {
							fileAreaMap := fileArea.(map[string]interface{})
							files := fileAreaMap["files"].([]interface{})
							for _, file := range files {
								fileMap := file.(map[string]interface{})
								moodleFile := MoodleFile{
									FileName: fileMap["filename"].(string),
									FileSize: int64(fileMap["filesize"].(float64)),
									FileURL:  fileMap["fileurl"].(string),
								}
								submissionMoodleFiles = append(submissionMoodleFiles, moodleFile)
							}
						}
					}
				}
			}
		}
	}

	var submissionMoodleFileNames []string
	for _, file := range submissionMoodleFiles {
		submissionMoodleFileNames = append(submissionMoodleFileNames, file.FileName)
	}

	var courseAssignment = courseApi.getCourseModAssignment(module)
	var introMoodleFileNames []string
	for _, file := range courseAssignment.IntroAttachment {
		introMoodleFileNames = append(introMoodleFileNames, file.FileName)
	}

	var data DownloadAssignmentData
	data.ID = module.ID
	data.CMID = module.ComponentID
	data.Name = module.Name
	data.ModName = module.ModName
	data.SubmissionAttachmentsNames = submissionMoodleFileNames
	data.Intro = courseAssignment.Intro
	data.SubmissionStatement = courseAssignment.SubmissionStatement
	data.IntroAttachmentsNames = introMoodleFileNames

	modulePath := fmt.Sprintf("%s/%s(%d)", basePath, module.Name, module.ID)
	introFilesPath := fmt.Sprintf("%s/introfiles", modulePath)
	submissionFilesPath := fmt.Sprintf("%s/submissions", modulePath)

	err = util.SaveStructToJSON(data, modulePath+"/assignment.json")
	if err != nil {
		return err
	}

	for _, file := range submissionMoodleFiles {
		err := courseApi.client.downloadFile(file.FileURL, submissionFilesPath+"/"+file.FileName, file.FileSize)
		if err != nil {
			return err
		}
	}
	for _, file := range courseAssignment.IntroAttachment {
		err := courseApi.client.downloadFile(file.FileURL, introFilesPath+"/"+file.FileName, file.FileSize)
		if err != nil {
			return err
		}
	}
	return nil

}

func (courseApi *CourseApi) getCourseModAssignment(module *CourseModule) *CourseModAssignment {
	for _, assignment := range courseApi.courseCache.CourseModAssignments {
		if assignment.AssignmentID == module.ID {
			return &assignment
		}
	}

	return nil
}
