package moodle

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"html/template"
	"os"
	"path/filepath"
	"scar/util"
)

type CoursesOverviewPage struct {
	Courses []CourseCard
}

type CourseCard struct {
	FullName        string        `json:"fullname"`
	ShortName       string        `json:"shortname"`
	Summary         template.HTML `json:"summary"`
	Category        string        `json:"coursecategory"`
	CourseImage     string        `json:"courseimage"`
	CourseImageType string        `json:"courseimagetype"`
	CourseImageURL  template.URL
}

type CoursePage struct {
}

func createMoodleWebsite() error {
	var archiverPath = util.Config.GetString("save_path")
	var moodlePath = filepath.Join(archiverPath, "moodle")
	if _, err := os.Stat(moodlePath); errors.Is(err, os.ErrNotExist) {
		return err
	}

	coursesOverviewPage, err := getCoursePage(moodlePath)
	if err != nil {
		logrus.Error("Could not retrieve Courses from moodle folder")
		return err
	}

	tmpl, err := template.ParseFiles("html/templates/moodle-courses-page.html")
	if err != nil {
		logrus.Fatal("Error loading template: ", err)
		return err
	}

	var outputPath = filepath.Join(archiverPath, "html", "moodle", "index.html")
	err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, coursesOverviewPage)
	if err != nil {
		return err
	}
	return nil
}

func getCoursePage(moodlePath string) (CoursesOverviewPage, error) {
	entries, err := os.ReadDir(moodlePath)
	if err != nil {
		return CoursesOverviewPage{}, err
	}
	var coursesPageData CoursesOverviewPage
	coursesPageData.Courses = []CourseCard{}
	for _, entry := range entries {
		if entry.IsDir() {
			file := filepath.Join(moodlePath, entry.Name(), "data.json")
			var data, err = os.ReadFile(file)
			if err != nil {
				logrus.Info("Could not get Course data for Course ", entry.Name(), ". Skipping it")
				continue
			}
			var courseData CourseCard
			err = json.Unmarshal(data, &courseData)
			if err != nil {
				logrus.Info("Could not parse Course data for Course ", entry.Name(), ". Skipping it")
				continue
			}
			courseData.CourseImageURL = template.URL(courseData.CourseImage)
			coursesPageData.Courses = append(coursesPageData.Courses, courseData)
		}
	}

	return coursesPageData, nil
}
