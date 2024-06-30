package moodle

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"scar/util"
	"strconv"
	"strings"
)

type coursesOverviewPage struct {
	Courses []courseData
}
type courseModule struct {
	ID      int
	Name    string
	ModName string
	Data    map[string]interface{}
	Path    string
}
type courseSection struct {
	Name          string `json:"name"`
	ID            int    `json:"id"`
	CourseModules []courseModule
}

type courseData struct {
	ID              int             `json:"id"`
	FullName        string          `json:"fullname"`
	ShortName       string          `json:"shortname"`
	Summary         template.HTML   `json:"summary"`
	Category        string          `json:"coursecategory"`
	CourseImage     string          `json:"courseimage"`
	CourseImageType string          `json:"courseimagetype"`
	Sections        []courseSection `json:"sections"`
	CourseImageURL  template.URL
}

type assignmentMod struct {
	ID                         int           `json:"id"`
	CMID                       int           `json:"cmid"`
	Name                       string        `json:"name"`
	Intro                      template.HTML `json:"intro"`
	Modname                    string        `json:"modname"`
	SubmissionStatement        template.HTML `json:"submissionstatement"`
	IntroAttachments           []string      `json:"introattachmentsnames"`
	SubmissionAttachments      []string      `json:"submissionattachmentsnames"`
	IntroAttachmentsPaths      []string
	SubmissionAttachmentsPaths []string
}
type labelMod struct {
	ID          int           `json:"id"`
	CMID        int           `json:"cmid"`
	Name        string        `json:"name"`
	Modname     string        `json:"modname"`
	Description template.HTML `json:"description"`
}

type resourceMod struct {
	Name             string   `json:"name"`
	ContentFileNames []string `json:"contentfilenames"`
	ContentFilePath  string
}
type urlMod struct {
	Name        string   `json:"name"`
	ContentUrls []string `json:"contenturls"`
}

func createMoodleWebsite() error {
	var archiverPath = util.Config.GetString("save_path")
	var moodlePath = filepath.Join(archiverPath, "moodle")
	if _, err := os.Stat(moodlePath); errors.Is(err, os.ErrNotExist) {
		return err
	}
	page, err := getCoursePage(moodlePath)
	if err != nil {
		logrus.Error("Could not retrieve Courses from moodle folder")
		return err
	}
	err = createOverviewPage(&page, archiverPath)
	if err != nil {
		logrus.Error("Could not create overview page")
		return err
	}

	for _, course := range page.Courses {
		err := createCoursePage(course, archiverPath)
		if err != nil {
			logrus.Error("Could not create Course page")
			continue
		}
	}

	if err := os.Symlink(moodlePath, filepath.Join(archiverPath, "html", "moodle", "data")); !errors.Is(err, os.ErrExist) {
		return err
	}
	return nil
}

func createOverviewPage(page *coursesOverviewPage, archiverPath string) error {
	var outputPath = filepath.Join(archiverPath, "html", "moodle", "index.html")
	tmpl, err := template.ParseFiles("html/templates/moodle-courses-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}
	err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, page)
	if err != nil {
		return err
	}
	return nil
}
func createCoursePage(course courseData, archiverPath string) error {
	var outputPath = filepath.Join(archiverPath, "html", "moodle", fmt.Sprintf("%d", course.ID), "index.html")
	tmpl, err := template.ParseFiles("html/templates/course/moodle-course-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}
	err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, course)
	if err != nil {
		return err
	}

	for _, section := range course.Sections {
		for _, module := range section.CourseModules {
			err := createModulePage(module, filepath.Dir(outputPath))
			if err != nil {
				logrus.Error("Could not create Module page:", err)
				continue
			}
		}
	}

	return nil
}
func createModulePage(mod courseModule, coursePath string) error {
	var outputPath = filepath.Join(coursePath, fmt.Sprintf("%d.html", mod.ID))
	assignTemp, err := template.ParseFiles("html/templates/course/mod/mod-assignment-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}
	labelTemp, err := template.ParseFiles("html/templates/course/mod/mod-label-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}
	resourceTemp, err := template.ParseFiles("html/templates/course/mod/mod-resource-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}
	urlTemp, err := template.ParseFiles("html/templates/course/mod/mod-url-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}

	err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	if mod.ModName == "assign" {
		jsonData, err := json.Marshal(mod.Data)
		if err != nil {
			return err
		}
		var assignment assignmentMod
		if err := json.Unmarshal(jsonData, &assignment); err != nil {
			return err
		}
		for _, attachment := range assignment.IntroAttachments {
			assignment.IntroAttachmentsPaths = append(assignment.IntroAttachmentsPaths, filepath.Join(mod.Path, "introfiles", attachment))
		}
		for _, attachment := range assignment.SubmissionAttachments {
			assignment.SubmissionAttachmentsPaths = append(assignment.SubmissionAttachmentsPaths, filepath.Join(mod.Path, "submissions", attachment))
		}
		err = assignTemp.Execute(outputFile, assignment)
		if err != nil {
			return err
		}
	} else if mod.ModName == "label" {
		jsonData, err := json.Marshal(mod.Data)
		if err != nil {
			return err
		}
		var label labelMod
		if err := json.Unmarshal(jsonData, &label); err != nil {
			return err
		}
		err = labelTemp.Execute(outputFile, label)
		if err != nil {
			return err
		}
	} else if mod.ModName == "resource" {
		jsonData, err := json.Marshal(mod.Data)
		if err != nil {
			return err
		}
		var resource resourceMod
		if err := json.Unmarshal(jsonData, &resource); err != nil {
			return err
		}
		resource.ContentFilePath = filepath.Join(mod.Path, "contents", resource.ContentFileNames[0])
		err = resourceTemp.Execute(outputFile, resource)
		if err != nil {
			return err
		}
	} else if mod.ModName == "url" {
		jsonData, err := json.Marshal(mod.Data)
		if err != nil {
			return err
		}
		var url urlMod
		if err := json.Unmarshal(jsonData, &url); err != nil {
			return err
		}
		err = urlTemp.Execute(outputFile, url)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Mod type is not supported: " + mod.ModName)
	}
	return nil
}
func getCoursePage(moodlePath string) (coursesOverviewPage, error) {
	entries, err := os.ReadDir(moodlePath)
	if err != nil {
		return coursesOverviewPage{}, err
	}
	var coursesPageData coursesOverviewPage
	coursesPageData.Courses = []courseData{}
	for _, entry := range entries {
		if entry.IsDir() {
			file := filepath.Join(moodlePath, entry.Name(), "data.json")
			var data, err = os.ReadFile(file)
			if err != nil {
				logrus.Info("Could not get Course data for Course ", entry.Name(), ". Skipping it")
				continue
			}
			var courseData courseData
			err = json.Unmarshal(data, &courseData)
			if err != nil {
				logrus.Info("Could not parse Course data for Course ", entry.Name(), ". Skipping it")
				continue
			}
			courseData.CourseImageURL = template.URL(courseData.CourseImage)
			sectionNameMap := make(map[int]string)
			for _, section := range courseData.Sections {
				sectionNameMap[section.ID] = section.Name
			}
			sections, err := getSections(filepath.Join(moodlePath, entry.Name()), sectionNameMap)
			if err != nil {
				logrus.Info("Could not retrieve sections. ", err)
			}
			courseData.Sections = sections
			coursesPageData.Courses = append(coursesPageData.Courses, courseData)
		}
	}

	return coursesPageData, nil
}

func getSections(coursePath string, sectionNameMap map[int]string) ([]courseSection, error) {
	entries, err := os.ReadDir(coursePath)
	if err != nil {
		return []courseSection{}, err
	}

	var sections []courseSection
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		var sec, err = getSection(filepath.Join(coursePath, entry.Name()))
		if err != nil {
			logrus.Info("Could not get section ", entry.Name(), ". Skipping it")
			continue
		}
		num, _ := strconv.Atoi(sec.Name)
		sec.Name = sectionNameMap[num]
		sections = append(sections, sec)
	}
	return sections, nil
}
func getSection(sectionPath string) (courseSection, error) {
	entries, err := os.ReadDir(sectionPath)
	if err != nil {
		return courseSection{}, err
	}

	var moodlePath = filepath.Join(util.Config.GetString("save_path"), "moodle")

	var section courseSection
	section.CourseModules = []courseModule{}
	_, section.Name = filepath.Split(sectionPath)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		var data, err = os.ReadFile(filepath.Join(sectionPath, entry.Name(), "data.json"))
		if err != nil {
			logrus.Info("Could not get Course data for Course ", entry.Name(), ". Because", err)
			continue
		}
		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}

		var mod courseModule
		mod.Name = result["name"].(string)
		mod.ModName = result["modname"].(string)
		mod.ID = int(result["id"].(float64))
		mod.Path = strings.Replace(filepath.Join(sectionPath, entry.Name()), moodlePath, "data", 1000)
		mod.Data = result
		section.CourseModules = append(section.CourseModules, mod)
	}

	return section, nil
}
