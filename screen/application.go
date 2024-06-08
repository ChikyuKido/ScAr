package screen

var App *ScreenManager

func CreateApplication() {
	App = NewScreenManager()
}
func AddScreen(screen *Screen) {
	App.AddScreen(screen)
}
func RunApplication() {
	App.BuildMainScreen()
	App.Run()
}
