package moodle

import (
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"scar/util"
)

type CoursesPage struct {
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

func CreateMoodleWebsite() error {
	var archiverPath = util.Config.GetString("save_path")
	var moodlePath = filepath.Join(archiverPath, "moodle")
	if _, err := os.Stat(moodlePath); errors.Is(err, os.ErrNotExist) {
		return err
	}

	entries, err := os.ReadDir(moodlePath)
	if err != nil {
		return err
	}
	var coursesPageData CoursesPage
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

	tmpl, err := template.ParseFiles("html/templates/moodle-courses-page.html")
	if err != nil {
		log.Fatal("Error loading template: ", err)
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

	err = tmpl.Execute(outputFile, coursesPageData)
	if err != nil {
		return err
	}

	cssBytes, err := os.ReadFile("html/css/bulma.css")
	if err != nil {
		logrus.Error("Could not get bulma css")
		return err
	}
	err = os.MkdirAll(filepath.Join(archiverPath, "html", "moodle", "css"), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(archiverPath, "html", "moodle", "css", "bulma.css"), cssBytes, os.ModePerm)
	if err != nil {
		logrus.Error("Could not write bulma css")
		return err
	}
	return nil
}
