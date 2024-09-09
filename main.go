package main

import (
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"scar/digi4school"
	"scar/moodle"
	"scar/screen"
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
	screen.CreateApplication()
	screen.AddScreen(moodle.GetMoodleScreen())
	screen.AddScreen(digi4school.GetD4SScreen())
	screen.RunApplication()

}
