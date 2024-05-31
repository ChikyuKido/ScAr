package main

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"scar/moodle"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	client := moodle.NewMoodleClient(true)
	client.ServiceUrl = os.Getenv("serviceUrl")
	err = client.Login(os.Getenv("username"), os.Getenv("password"))
	if err != nil {
		logrus.Fatal("Failed to login: ", err.Error())
	}

	courses, err := client.CourseApi.GetCourses(false)

	if err != nil {
		logrus.Fatal("Failed to get courses: ", err.Error())
	}

	client.CourseApi.FetchCourseContents(&courses[0])
	for _, module := range moodle.GetAllAssignModules(&courses[0]) {
		err = client.CourseApi.DownloadAssignModule(&module, "archiver/moodle/"+courses[0].ShortName)
		if err != nil {
			logrus.Fatal("Could not retrieve course contents: ", err.Error())
		}
	}

	/**mathCourse := courses[0]
	err = client.CourseApi.FetchCourseContents(&mathCourse)
	if err != nil {
		logrus.Fatal("Could not retrieve course contents: ", err.Error())
	}

	err = client.CourseApi.DownloadAssignModule(&moodle.GetAllAssignModules(&mathCourse)[0], "archiver/moodle/pos")
	if err != nil {
		logrus.Fatal("Could not retrieve course contents: ", err.Error())
	}*/

	/**
	app := tview.NewApplication()
	list := tview.NewList()
	var screens = []util.Screen{moodle.GetMoodleScreen(app, list)}
	for i, screen := range screens {
		list.AddItem(screen.Name, "", rune(i+1+'0'), func() {
			app.SetRoot(screen.Root, true)
		})
	}
	list.SetTitle("SCAR").SetBorder(true)

	app.SetFocus(list)
	if err := app.SetRoot(list, true).Run(); err != nil {
		panic(err)
	}*/
}
