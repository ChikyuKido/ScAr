package digi4school

import (
	"fmt"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"scar/screen"
	"scar/util"
	"strings"
	"time"
)

type digi4SchoolContext struct {
	digi4s *Digi4SchoolClient
	books  []Book
}

var digi4sContext = digi4SchoolContext{}

func GetD4SScreen() *screen.Screen {
	username := util.Config.GetStringWD("digi4s_username", "")
	password := util.Config.GetStringWD("digi4s_password", "")
	digi4sContext.digi4s = NewDigi4SClient(username, password)
	return &screen.Screen{Name: "Digi4School",
		DownloadPage: getStartFunc(),
		CreateHtml: func() error {
			return nil
		},
		FolderName: "d4s",
		ImageName:  "d4s.png"}
}

func getStartFunc() *tview.List {
	var list = tview.NewList()
	list.SetTitle("ScAr - Digi4School")
	list.SetBorder(true)
	list.SetFocusFunc(func() {
		if digi4sContext.digi4s.Login() != nil {
			screen.App.ShowPopup("Wrong credentials. Please check the config file.", screen.App.MainScreen, screen.App.MainScreen)
		} else {
			screen.App.SwitchScreen(getBookList())
		}
	})
	return list
}

func getBookList() *tview.List {
	var list = tview.NewList()
	list.SetTitle("ScAr - Digi4School")
	list.SetBorder(true)
	books, _ := digi4sContext.digi4s.GetBooks()
	digi4sContext.books = books
	if len(books) == 0 {
		list.AddItem("You dont have any books", "Nooooo boooks", 0, nil)
		return list
	}
	list.AddItem("0) All", "All Courses", 0, func() {
		screen.App.SwitchScreen(GetProgressView(-1))
	})
	for i, book := range books {
		list.AddItem(fmt.Sprintf("%d) %s", i+1, book.Name),
			fmt.Sprintf("DataId: %s DataCode: %s", book.DataId, book.DataCode), 0, func() {
				screen.App.SwitchScreen(GetProgressView(i))
			})
	}
	list.AddItem(fmt.Sprintf("%d) Back", len(books)+1), "", 0, func() {
		screen.App.SwitchToMainScreen()
	})

	return list
}

func GetProgressView(index int) tview.Primitive {
	var coursesSize = 1
	var booksToDownload []Book
	if index == -1 {
		coursesSize = len(digi4sContext.books)
		booksToDownload = digi4sContext.books
	} else {
		booksToDownload = []Book{digi4sContext.books[index]}
	}
	bookProgressBar := tview.NewTextView().SetScrollable(false)
	pageProgressBar := tview.NewTextView().SetScrollable(false).
		SetChangedFunc(func() {
			screen.App.App.Draw()
		})
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
	bookChan := make(chan int, coursesSize)
	pageChan := make(chan int, 999)
	go func() {
		for running {
			bookTotal := cap(bookChan)
			pageTotal := cap(pageChan)
			pageProgressCount := len(pageChan)
			bookProgressCount := len(bookChan)
			pageProgress := (pageProgressCount) * 100 / pageTotal
			bookProgress := (bookProgressCount) * 100 / bookTotal
			pageProgressBar.SetText(fmt.Sprintf("Page [%d|%d]: [%-50s] %d%%", pageProgressCount, pageTotal, strings.Repeat("=", pageProgress/2), pageProgress))
			bookProgressBar.SetText(fmt.Sprintf("Book [%d|%d]: [%-50s] %d%%", bookProgressCount, bookTotal, strings.Repeat("=", bookProgress/2), bookProgress))
			time.Sleep(10 * time.Millisecond)
		}
	}()
	go func() {
		for i, book := range booksToDownload {
			var basePath = util.Config.GetString("save_path")
			err := digi4sContext.digi4s.DownloadBook(&book, filepath.Join(basePath, "digi4s"), pageChan)
			if err != nil {
				logrus.Error("Failed to download Course: ", err.Error())
				continue
			}
			bookChan <- i + 1
		}
		screen.App.SwitchScreen(getBookList())
		close(pageChan)
		close(bookChan)
		running = false
		digi4sContext.digi4s.Logout()
	}()
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(bookProgressBar, 0, 1, false).
			AddItem(pageProgressBar, 0, 1, false), 0, 1, false).
		AddItem(logTextView, 0, 8, true)
	return flex
}
