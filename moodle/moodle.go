package moodle

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"os"
	"scar/util"
	"strings"
	"time"
)

type MoodleCache struct {
	course []Course
}

var moodleCache MoodleCache
var moodleClient = NewMoodleClient(true)

func GetMoodleScreen(app *tview.Application, mainScreen tview.Primitive) util.Screen {
	return util.Screen{Name: "Moodle", Root: GetStartList(app, mainScreen)}
}

func GetStartList(app *tview.Application, mainScreen tview.Primitive) *tview.List {
	var list = tview.NewList()
	list.SetTitle("Moodle")
	list.SetBorder(true)
	list.AddItem("Download", "", '1', func() {
		if len(moodleClient.Token) == 0 {
			moodleClient.ServiceUrl = os.Getenv("serviceUrl")
			moodleClient.Login(os.Getenv("username"), os.Getenv("password"))
			//	app.SetRoot(GetPasswordDialogModal(app, mainScreen), true)
			//return
		}
		app.SetRoot(GetDownloadView(app, mainScreen), true)
	})
	list.AddItem("Back", "", '2', func() {
		app.SetRoot(mainScreen, true)
	})
	return list
}

func GetDownloadView(app *tview.Application, mainScreen tview.Primitive) *tview.List {
	if len(moodleCache.course) == 0 {
		courses, err := moodleClient.CourseApi.GetCourses(false)
		moodleCache.course = courses
		if err != nil {
			return nil
		}
	}
	var list = tview.NewList()
	list.AddItem("0) All", "All Courses", 0, func() {
		app.SetRoot(GetProgressView(app, mainScreen, -1), true)
	})
	for i, course := range moodleCache.course {
		list.AddItem(fmt.Sprintf("%d) %s", i+1, course.ShortName),
			fmt.Sprintf("%s (%d)", course.Fullname, course.ID), 0, func() {
				app.SetRoot(GetProgressView(app, mainScreen, i), true)
			})
	}
	list.AddItem(fmt.Sprintf("%d) Back", len(moodleCache.course)+1), "", 0, func() {
		app.SetRoot(GetStartList(app, mainScreen), true)
	})

	return list
}

func GetPasswordDialogModal(app *tview.Application, mainScreen tview.Primitive) tview.Primitive {
	serviceUrl := tview.NewInputField().
		SetLabel("Moodle URL: ").
		SetFieldWidth(0)
	usernameInput := tview.NewInputField().
		SetLabel("Username: ").
		SetFieldWidth(0)

	passwordInput := tview.NewInputField().
		SetLabel("Password: ").
		SetFieldWidth(0).
		SetMaskCharacter('*')
	serviceUrl.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(usernameInput)
	})
	usernameInput.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(passwordInput)
	})
	passwordInput.SetDoneFunc(func(key tcell.Key) {
		moodleClient.ServiceUrl = serviceUrl.GetText()
		err := moodleClient.Login(usernameInput.GetText(), passwordInput.GetText())
		if err != nil {
			modal := tview.NewModal().
				SetText("Wrong credentials or service url").
				AddButtons([]string{"Cancel", "Ok"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Cancel" {
						app.SetRoot(GetStartList(app, mainScreen), true)
					} else {
						app.SetRoot(GetPasswordDialogModal(app, mainScreen), true)
					}
				})
			app.SetRoot(modal, true)
			return
		}
		app.SetRoot(GetDownloadView(app, mainScreen), true)
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(serviceUrl, 0, 1, true).
		AddItem(usernameInput, 0, 1, false).
		AddItem(passwordInput, 0, 1, false)
	flex.SetBorder(true).SetTitle("Credentials")

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyUp {
			app.SetFocus(usernameInput)
			return nil
		} else if event.Key() == tcell.KeyDown {
			app.SetFocus(passwordInput)
			return nil
		}
		return event
	})
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}

	return modal(flex, 40, 10)
}

func GetProgressView(app *tview.Application, mainScreen tview.Primitive, index int) tview.Primitive {
	var coursesSize = 1
	var coursesToDownload []Course
	if index == -1 {
		coursesSize = len(moodleCache.course)
		coursesToDownload = moodleCache.course
	} else {
		coursesToDownload = []Course{moodleCache.course[index]}
	}
	courseProgressBar := tview.NewTextView().SetScrollable(false)
	modulesProgressBar := tview.NewTextView().SetScrollable(false)
	logTextView := tview.NewTextView().
		SetChangedFunc(func() {
			app.Draw()
		})
	logTextView.SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetBorder(true).
		SetTitle("Log Output")
	var running = true
	coursesChan := make(chan int, coursesSize)
	modulesChan := make(chan int, 1)
	go func() {
		for running {
			modulesTotal := cap(modulesChan)
			coursesTotal := cap(coursesChan)
			modulesProgressCount := len(modulesChan)
			coursesProgressCount := len(coursesChan)
			modulesProgress := (modulesProgressCount) * 100 / modulesTotal
			coursesProgress := (coursesProgressCount) * 100 / coursesTotal
			modulesProgressBar.SetText(fmt.Sprintf("Module [%d|%d]: [%-50s] %d%%", modulesProgressCount, modulesTotal, strings.Repeat("=", modulesProgress/2), modulesProgress))
			courseProgressBar.SetText(fmt.Sprintf("Course [%d|%d]: [%-50s] %d%%", coursesProgressCount, coursesTotal, strings.Repeat("=", coursesProgress/2), coursesProgress))
			time.Sleep(10 * time.Millisecond)
		}
	}()
	go func() {
		for i, course := range coursesToDownload {
			err := moodleClient.CourseApi.FetchCourseContents(&course)
			if err != nil {
				logrus.Error("Failed to download Course: ", err.Error())
				continue
			}
			modulesChan = make(chan int, len(GetAllModules(&course)))

			err = moodleClient.CourseApi.DownloadCourse(&course, "/home/kido/archiver/moodle", modulesChan, logTextView)
			if err != nil {
				logrus.Error("Failed to download Course: ", err.Error())
				continue
			}

			coursesChan <- i + 1
		}
		app.SetRoot(GetDownloadView(app, mainScreen), true)
		running = false
	}()
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(courseProgressBar, 0, 1, false).
			AddItem(modulesProgressBar, 0, 1, false), 0, 1, false).
		AddItem(logTextView, 0, 8, true)
	return flex
}
