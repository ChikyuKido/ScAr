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
)

type coursesOverviewPage struct {
	Courses []courseData
}
type courseModule struct {
	Name    string
	ModName string
	Data    map[string]interface{}
}
type courseSection struct {
	Name          string
	CourseModules []courseModule
}

type courseData struct {
	FullName        string        `json:"fullname"`
	ShortName       string        `json:"shortname"`
	Summary         template.HTML `json:"summary"`
	Category        string        `json:"coursecategory"`
	CourseImage     string        `json:"courseimage"`
	CourseImageType string        `json:"courseimagetype"`
	Sections        []courseSection
	CourseImageURL  template.URL
}

type assignmentMod struct {
	ID                    int           `json:"id"`
	CMID                  int           `json:"cmid"`
	Name                  string        `json:"name"`
	Intro                 template.HTML `json:"intro"`
	Modname               string        `json:"modname"`
	SubmissionStatement   template.HTML `json:"submissionstatement"`
	IntroAttachments      []string      `json:"introattachmentsnames"`
	SubmissionAttachments []string      `json:"submissionattachmentsnames"`
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
	var outputPath = filepath.Join(archiverPath, "html", "moodle", course.ShortName, "index.html")
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
	var outputPath = filepath.Join(coursePath, mod.Name+".html")
	assignTemp, err := template.ParseFiles("html/templates/course/mod/mod-assignment-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}
	//TODO: other module types

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
		err = assignTemp.Execute(outputFile, assignment)
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
			sections, err := getSections(filepath.Join(moodlePath, entry.Name()))
			if err != nil {
				logrus.Info("Could not retrieve sections. ", err)
			}
			courseData.Sections = sections
			coursesPageData.Courses = append(coursesPageData.Courses, courseData)
		}
	}

	return coursesPageData, nil
}

func getSections(coursePath string) ([]courseSection, error) {
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
		sections = append(sections, sec)
	}
	return sections, nil
}
func getSection(sectionPath string) (courseSection, error) {
	entries, err := os.ReadDir(sectionPath)
	if err != nil {
		return courseSection{}, err
	}

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
		mod.Data = result
		section.CourseModules = append(section.CourseModules, mod)
	}

	return section, nil
}
