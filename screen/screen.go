package screen

import (
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

type ScreenManager struct {
	App        *tview.Application
	MainScreen tview.Primitive
	Screens    []*Screen
}

type Screen struct {
	Name         string
	DownloadPage tview.Primitive
	CreateHtml   func() error
	ImageName    string
	FolderName   string
}

func NewScreenManager() *ScreenManager {
	return &ScreenManager{
		App: tview.NewApplication(),
	}
}
func (sm *ScreenManager) AddScreen(screen *Screen) {
	sm.Screens = append(sm.Screens, screen)
}

func (sm *ScreenManager) BuildMainScreen() {
	mainList := tview.NewList().
		AddItem("Download", "Download content from a specific provider", '1', func() {
			downloadList := tview.NewList()
			for i, screen := range sm.Screens {
				downloadList.AddItem(screen.Name, "", rune('1'+i), func() {
					sm.App.SetRoot(screen.DownloadPage, true)
				})
			}
			downloadList.AddItem("Back", "", 'b', func() {
				sm.SwitchToMainScreen()
			})
			downloadList.SetTitle("ScAr - Download").SetBorder(true)
			sm.SwitchScreen(downloadList)
		}).AddItem("Create HTML", "Creates a html for a specific provider", '2', func() {
		htmlList := tview.NewList()
		htmlList.AddItem("All", "", '0', func() {
			err := createIndexHtml()
			if err != nil {
				logrus.Errorf("Could not create index html. Because: %s", err.Error())
			}
			for _, screen := range sm.Screens {
				err = screen.CreateHtml()
				if err != nil {
					logrus.Errorf("Could not create html for %s. Because: %s", screen.Name, err.Error())
				}
			}
		})
		for i, screen := range sm.Screens {
			htmlList.AddItem(screen.Name, "", rune('1'+i), func() {
				err := createIndexHtml()
				if err != nil {
					logrus.Errorf("Could not create index html. Because: %s", err.Error())
				}
				err = screen.CreateHtml()
				if err != nil {
					logrus.Errorf("Could not create html for %s. Because: %s", screen.Name, err.Error())
				}
			})
		}
		htmlList.AddItem("Back", "", 'b', func() {
			sm.SwitchScreen(sm.MainScreen)
		})
		htmlList.SetTitle("ScAr - Download").SetBorder(true)
		sm.SwitchScreen(htmlList)
	})
	mainList.SetTitle("ScAr").SetBorder(true)
	sm.MainScreen = mainList
}

func (sm *ScreenManager) SwitchScreen(root tview.Primitive) {

	sm.App.SetRoot(root, true)
}

func (sm *ScreenManager) SwitchToMainScreen() {
	sm.App.SetRoot(sm.MainScreen, true)
}

func (sm *ScreenManager) SetFocus(root tview.Primitive) {
	sm.App.SetFocus(root)
}
func (sm *ScreenManager) Run() {

	if err := sm.App.SetRoot(sm.MainScreen, true).Run(); err != nil {
		panic(err)
	}
}
