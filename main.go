package main

import (
	"github.com/joho/godotenv"
	"github.com/rivo/tview"
	"scar/moodle"
	"scar/util"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
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
	}
}
