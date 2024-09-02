package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"scar/digi4school"
	"scar/moodle"
	"scar/screen"
	"scar/util"
)

func main() {

	client := digi4school.NewDigi4SClient(email, password)

	if err := client.Login(); err == nil {
		client.GetBookCookie("23s5agvhgkxf")
		client.Logout()
	} else {
		fmt.Println("Login failed")
	}
	return
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
