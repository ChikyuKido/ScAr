package digi4school

import (
	"github.com/rivo/tview"
	"scar/screen"
)

func GetDigi4SchoolScreen(app *tview.Application, mainScreen tview.Primitive) screen.Screen {
	return screen.Screen{Name: "Digi4School", Root: GetStartList(app, mainScreen)}
}

func GetStartList(app *tview.Application, mainScreen tview.Primitive) *tview.List {
	var list = tview.NewList()
	list.SetTitle("Moodle")
	list.SetBorder(true)
	list.AddItem("Download", "", '1', func() {

	})
	list.AddItem("Back", "", '2', func() {
		app.SetRoot(mainScreen, true)
	})
	return list
}
