package screen

import (
	"github.com/sirupsen/logrus"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"scar/util"
)

type ProviderPage struct {
	Provider []ProviderCard
}

type ProviderCard struct {
	Name       string
	ImageName  string
	FolderName string
}

func createIndexHtml() error {
	var archiverPath = util.Config.GetString("save_path")

	tmpl, err := template.ParseFiles("html/index.html")
	if err != nil {
		log.Fatal("Error loading template: ", err)
	}
	var providerPage ProviderPage
	for _, screen := range App.Screens {
		var pc ProviderCard
		pc.Name = screen.Name
		pc.ImageName = screen.ImageName
		pc.FolderName = screen.FolderName
		providerPage.Provider = append(providerPage.Provider, pc)
	}

	var outputPath = filepath.Join(archiverPath, "html", "index.html")
	err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return err
	}
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, providerPage)
	if err != nil {
		return err
	}

	cssBytes, err := os.ReadFile("html/css/bulma.css")
	if err != nil {
		logrus.Error("Could not get bulma css")
		return err
	}
	var bulmaPath = filepath.Join(archiverPath, "html", "css", "bulma.css")
	err = os.MkdirAll(filepath.Dir(bulmaPath), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(bulmaPath, cssBytes, os.ModePerm)
	if err != nil {
		logrus.Error("Could not write bulma css")
		return err
	}
	return nil
}
