package digi4school

import (
	"fmt"
	"github.com/rivo/tview"
	"scar/screen"
	"scar/util"
)

type digi4SchoolContext struct {
	digi4s *Digi4SchoolClient
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
			modal := tview.NewModal().
				SetText("Wrong credentials. Please check the config file.").
				AddButtons([]string{"Cancel", "Ok"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonLabel == "Cancel" {
						screen.App.SwitchToMainScreen()
					} else {
						screen.App.SwitchToMainScreen()
					}
				})
			screen.App.SwitchScreen(modal)
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
	if len(books) == 0 {
		list.AddItem("You dont have any books", "Nooooo boooks", 0, nil)
		return list
	}
	list.AddItem("0) All", "All Courses", 0, func() {
		//screen.App.SwitchScreen(GetProgressView(-1))
	})
	for i, book := range books {
		list.AddItem(fmt.Sprintf("%d) %s", i+1, book.Name),
			fmt.Sprintf("DataId: %s DataCode: %s", book.DataId, book.DataCode), 0, func() {
				//	screen.App.SwitchScreen(GetProgressView(i))
			})
	}
	list.AddItem(fmt.Sprintf("%d) Back", len(books)+1), "", 0, func() {
		screen.App.SwitchToMainScreen()
	})

	return list
}
