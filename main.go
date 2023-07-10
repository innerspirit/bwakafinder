package main

import (
	"fmt"
	"image/color"
	"io/ioutil"

	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/innerspirit/getscprocess/lib"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("BW Aka Finder")

	data := [][]string{
		{"AKA", "Max MMR", "Rank"},
	}

	myTable := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0]) //rows, columns
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Content") // basic cell template
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i.Row][i.Col]) // set content for cell
		},
	)

	bgColor := color.RGBA{20, 20, 20, 255}
	errorLabel := canvas.NewText("", color.RGBA{255, 0, 0, 255}) // red color error label for displaying errors
	errorLabel.TextStyle.Bold = true
	myWindow.SetContent(container.NewMax(canvas.NewRectangle(bgColor), container.NewMax(myTable), errorLabel))
	myWindow.Resize(fyne.NewSize(400, 400))

	opponent := "ZAELOT"

	dataChannel := make(chan [][]string)
	errorChannel := make(chan error)

	go func() {
		for {
			newData, err := grabPlayerInfo(opponent)
			if err != nil {
				errorChannel <- err
				time.Sleep(15 * time.Second)
				continue
			}
			dataChannel <- newData
			time.Sleep(15 * time.Second)
		}
	}()

	go func() {
		for newData := range dataChannel {
			data = newData
			myTable.Refresh()
			errorLabel.Text = "" // clear error label when new data is successfully fetched
		}
	}()

	go func() {
		for err := range errorChannel {
			errorLabel.Text = err.Error() // display error message in label
		}
	}()
	myWindow.ShowAndRun()
}

func grabPlayerInfo(player string) ([][]string, error) {
	_, port, err := lib.GetProcessInfo(false)
	fmt.Println("sc port: ", port)
	path := "aurora-profile-by-toon/"
	params := "/30?request_flags=scr_mmgameloading"
	url := fmt.Sprint("http://localhost:", port, "/web-api/v2/", path, player, params)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)

	if port == -1 {
		return nil, fmt.Errorf("SC:R is not running")
	}
	if err != nil {
		return nil, err
	} else {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		} else {
			_, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				return nil, readErr
			} else {
				return [][]string{
					{"AKA", "Max MMR", "Rank"},
					{"NewData1", "NewData2", "NewData3"},
				}, nil
			}
		}
	}
}
