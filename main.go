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
	err = client.Login(os.Getenv("username"), os.Getenv("password"))
	if err != nil {
		logrus.Fatal("Failed to login: ", err.Error())
	}

	courses, err := client.CourseApi.GetCourses(false)

	if err != nil {
		logrus.Fatal("Failed to get courses: ", err.Error())
	}

	mathCourse := courses[6]
	err = client.CourseApi.FetchCourseContents(&mathCourse)
	if err != nil {
		logrus.Fatal("Could not retrieve course contents: ", err.Error())
	}

	client.CourseApi.GetAssignModule(getAllAssignModules(&mathCourse)[0])
}
