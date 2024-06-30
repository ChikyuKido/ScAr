package moodle

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"scar/screen"
	"scar/util"
	"strings"
	"time"
)

type MoodleCache struct {
	course []Course
}

var moodleCache MoodleCache
var moodleClient = NewMoodleClient(true)

func GetMoodleScreen() *screen.Screen {
	return &screen.Screen{
		Name:         "Moodle",
		DownloadPage: getDownloadViewStart(),
		CreateHtml:   createMoodleWebsite,
		FolderName:   "moodle",
		ImageName:    "Moodle.png"}
}
func getDownloadViewStart() tview.Primitive {
	var box = tview.NewBox()
	box.SetFocusFunc(func() {
		if moodleClient.Token == "" {
			screen.App.SwitchScreen(GetPasswordDialogModal())
		} else {
			screen.App.SwitchScreen(GetDownloadView())
		}
	})
	return box
}
func GetDownloadView() *tview.List {
	if len(moodleCache.course) == 0 {
		courses, err := moodleClient.CourseApi.GetCourses(false)
		moodleCache.course = courses
		if err != nil {
			return nil
		}
	}
	var list = tview.NewList()
	list.AddItem("0) All", "All Courses", 0, func() {
		screen.App.SwitchScreen(GetProgressView(-1))
	})
	for i, course := range moodleCache.course {
		list.AddItem(fmt.Sprintf("%d) %s", i+1, course.ShortName),
			fmt.Sprintf("%s (%d)", course.Fullname, course.ID), 0, func() {
				screen.App.SwitchScreen(GetProgressView(i))
			})
	}
	list.AddItem(fmt.Sprintf("%d) Back", len(moodleCache.course)+1), "", 0, func() {
		screen.App.SwitchToMainScreen()
	})

	return list
}

func GetPasswordDialogModal() tview.Primitive {
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
		screen.App.SetFocus(usernameInput)
	})
	usernameInput.SetDoneFunc(func(key tcell.Key) {
		screen.App.SetFocus(passwordInput)
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
						screen.App.SwitchToMainScreen()
					} else {
						screen.App.SwitchScreen(GetPasswordDialogModal())
					}
				})
			screen.App.SwitchScreen(modal)
			return
		}
		screen.App.SwitchScreen(GetDownloadView())
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(serviceUrl, 0, 1, true).
		AddItem(usernameInput, 0, 1, false).
		AddItem(passwordInput, 0, 1, false)
	flex.SetBorder(true).SetTitle("Credentials")
	var currentIndex = 0
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown {
			currentIndex++
			if currentIndex > 2 {
				currentIndex = 0
			}
		} else if event.Key() == tcell.KeyUp {
			currentIndex--
			if currentIndex < 0 {
				currentIndex = 2
			}
		}
		if currentIndex == 0 {
			screen.App.SetFocus(serviceUrl)
		} else if currentIndex == 1 {
			screen.App.SetFocus(usernameInput)
		} else if currentIndex == 2 {
			screen.App.SetFocus(passwordInput)
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

func GetProgressView(index int) tview.Primitive {
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
			screen.App.App.Draw()
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

			var basePath = util.Config.GetString("save_path")
			err = moodleClient.CourseApi.DownloadCourse(&course, filepath.Join(basePath, "moodle"), modulesChan, logTextView)
			if err != nil {
				logrus.Error("Failed to download Course: ", err.Error())
				continue
			}

			coursesChan <- i + 1
		}
		screen.App.SwitchScreen(GetDownloadView())
		running = false
	}()
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(courseProgressBar, 0, 1, false).
			AddItem(modulesProgressBar, 0, 1, false), 0, 1, false).
		AddItem(logTextView, 0, 8, true)
	return flex
}
