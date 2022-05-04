package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func parse() {
	// read file
	inputFile, err := os.Open(openFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer inputFile.Close()
	fileScanner := bufio.NewScanner(inputFile)
	result := []Result{}
	part := ""
	for fileScanner.Scan() {
		row := fileScanner.Text()
		re := regexp.MustCompile(`Excited State +\d+`)
		if re.MatchString(row) {
			r := regexp.MustCompile(`f=\d+\.\d+`)

			fs := r.FindString(row)
			f, _ := strconv.ParseFloat(fs[3:], 64)
			if f > strTh*0.99 {
				part = row
			}
			continue
		}

		rer := regexp.MustCompile(`^\s+\d+ (->|<-)\d+\s+(\s|-)\d.\d+$`)
		if !rer.MatchString(row) {
			part = ""
			continue
		}

		re = regexp.MustCompile(`(-|)\d.\d+$`)
		if part != "" {
			v := re.FindString(row)
			if v == "" {
				continue
			}
			d, _ := strconv.ParseFloat(v, 64)
			if d > contTh {
				result = append(result, Result{part, row})
			}
		}

	}
	if err := fileScanner.Err(); err != nil {
		log.Fatalf("Error while reading file: %s", err)
	}

	outputFile, err := os.OpenFile(saveFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outputFile.Close()
	title := ""
	outputFile.WriteString(string(title))
	for i := 0; i < len(result); i++ {
		w := ""
		if i == 0 || i > 0 && result[i].Part != result[i-1].Part {
			w += "\n" + result[i].Part + "\n"
		}
		w += result[i].MatchRow + "\n"

		outputFile.WriteString(string(w))
	}

}

type Result struct {
	Part     string `json:"part name"`
	MatchRow string `json:"matched row"`
}

var saveFilePath string
var openFilePath string
var contTh float64
var strTh float64

func initApp() {
	if runtime.GOOS == "darwin" {
		os.Setenv("FYNE_FONT", `/System/Library/Fonts/Supplemental/Arial Unicode.ttf`)
	} else if runtime.GOOS == "windows" {
		os.Setenv("FYNE_FONT", `C:\Windows\Fonts\meiryo.ttc`)
	}
	openFilePath, _ = os.Getwd()
	saveFilePath, _ = os.Getwd()
	strTh = 0.02
	contTh = 0.09
}

func main() {
	initApp()

	a := app.New()
	w := a.NewWindow("Gaussian Parser")
	w.Resize(fyne.NewSize(640, 480))
	inputC := widget.NewEntry()
	inputS := widget.NewEntry()
	inputC.SetPlaceHolder("Enter contribution threshold...")
	inputS.SetPlaceHolder("Enter strength threshold...")

	openLabel := widget.NewLabel("Open file path:  " + openFilePath)
	saveLabel := widget.NewLabel("Save file path:  " + saveFilePath)

	selectFileButton := widget.NewButton("select file", func() {
		showDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err != nil {
				fmt.Println(err)
				return
			} else if file == nil {
				return
			}
			openFilePath = file.URI().Path()
			openLabel.SetText("Open file path:  " + openFilePath)
		}, w)
		showDialog.Show()
	})
	selectFileButton.Resize(fyne.NewSize(50, 0))
	saveFileButton := widget.NewButton("save file", func() {
		saveDialog := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
			if err != nil {
				fmt.Println(err)
				return
			} else if file == nil {
				return
			}
			saveFilePath = file.URI().Path()
			saveLabel.SetText("Save file path:  " + saveFilePath)
		}, w)
		saveDialog.Show()
	})
	openLabel.Move(fyne.NewPos(30, 0))
	saveLabel.Move(fyne.NewPos(30, 60))
	selectFileButton.Resize(fyne.NewSize(150, 30))
	saveFileButton.Resize(fyne.NewSize(150, 30))
	selectFileButton.Move(fyne.NewPos(400, 30))
	saveFileButton.Move(fyne.NewPos(400, 100))

	entryC := container.NewVBox(inputC, widget.NewButton("Set contribution threshold", func() {
		contTh, _ = strconv.ParseFloat(inputC.Text, 64)
	}))
	entryC.Move(fyne.NewPos(50, 150))
	entryS := container.NewVBox(inputS, widget.NewButton("Set strength threshold", func() {
		strTh, _ = strconv.ParseFloat(inputS.Text, 64)
	}))
	entryS.Move(fyne.NewPos(320, 150))

	parseButton := widget.NewButton("parse", func() {
		parse()
	})
	parseButton.Resize(fyne.NewSize(200, 50))
	parseButton.Move(fyne.NewPos(200, 270))

	w.SetContent(container.NewWithoutLayout(selectFileButton, saveFileButton, parseButton, entryC, entryS, saveLabel, openLabel))
	w.ShowAndRun()
}
