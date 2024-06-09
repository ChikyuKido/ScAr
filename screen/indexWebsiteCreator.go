package screen

import (
	"github.com/sirupsen/logrus"
	"html/template"
	"io"
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

	moveImgFolder("html/imgs", filepath.Join(archiverPath, "html", "imgs"))
	return nil
}

func moveImgFolder(srcDir, dstDir string) {
	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		logrus.Errorf("failed to create destination directory: %v", err)
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		logrus.Errorf("failed to read source directory: %v", err)
	}
	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if !entry.IsDir() {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				logrus.Warnf("failed to move %s to %s: %v", srcPath, dstPath, err)
			}
		}
	}
}

func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
