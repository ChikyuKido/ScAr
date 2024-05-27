package main

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	client := NewMoodleClient(os.Getenv("serviceUrl"), true)
	err = client.login(os.Getenv("username"), os.Getenv("password"))
	if err != nil {
		logrus.Fatal("Failed to login: ", err.Error())
	}

	courses, err := client.CourseApi.getCourses()

	if err != nil {
		logrus.Fatal("Failed to get courses: ", err.Error())
	}
	for _, course := range courses {
		logrus.Println(course.Fullname)
	}
}
