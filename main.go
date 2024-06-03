package main

import (
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"scar/digi4school"
	"scar/moodle"
	"scar/util"
)

func main() {
	file, err := os.OpenFile("logfile.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	logrus.SetOutput(file)

	util.Config.Load()

	app := tview.NewApplication()
	list := tview.NewList()
	var screens = []util.Screen{moodle.GetMoodleScreen(app, list), digi4school.GetDigi4SchoolScreen(app, list)}
	for i, screen := range screens {
		list.AddItem(screen.Name, "", rune(i+1+'0'), func() {
			app.SetRoot(screen.Root, true)
		})
	}
	list.SetTitle("ScAr").SetBorder(true)

	app.SetFocus(list)
	if err := app.SetRoot(list, true).Run(); err != nil {
		panic(err)
	}
}
