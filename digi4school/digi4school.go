package digi4school

import (
	"github.com/rivo/tview"
	"scar/screen"
)

func GetD4SScreen() *screen.Screen {
	return &screen.Screen{Name: "Digi4School",
		DownloadPage: GetStartList(),
		CreateHtml: func() error {
			return nil
		},
		FolderName: "d4s",
		ImageName:  "d4s.png"}
}

func GetStartList() *tview.List {
	var list = tview.NewList()
	list.SetTitle("ScAr - Digi4School")
	list.SetBorder(true)
	return list
}
