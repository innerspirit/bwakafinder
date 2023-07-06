package main

import (
	"fmt"
	// "image/color"
	"io/ioutil"

	"net/http"

	// "fyne.io/fyne/v2"
	// "fyne.io/fyne/v2/app"
	// "fyne.io/fyne/v2/canvas"
	// "fyne.io/fyne/v2/container"
	// "fyne.io/fyne/v2/widget"
	"github.com/innerspirit/getscprocess/lib"
)

func main() {
	// myApp := app.New()
	// myWindow := myApp.NewWindow("BW Aka Finder")

	// data := [][]string{
	// 	{"AKA", "Max MMR", "Rank"},
	// 	{"Data1", "Data2", "Data3"},
	// 	{"Data4", "Data5", "Data6"},
	// 	{"Data7", "Data8", "Data9"},
	// }

	// myTable := widget.NewTable(
	// 	func() (int, int) {
	// 		return len(data), len(data[0]) //rows, columns
	// 	},
	// 	func() fyne.CanvasObject {
	// 		return widget.NewLabel("Content") // basic cell template
	// 	},
	// 	func(i widget.TableCellID, o fyne.CanvasObject) {
	// 		o.(*widget.Label).SetText(data[i.Row][i.Col]) // set content for cell
	// 	},
	// )

	// bgColor := color.RGBA{20, 20, 20, 255}
	// myWindow.SetContent(container.NewMax(canvas.NewRectangle(bgColor), container.NewMax(myTable)))
	// myWindow.Resize(fyne.NewSize(400, 400))

	_, port, err := lib.GetProcessInfo(false)
	fmt.Println("sc port: ", port)
	path := "aurora-profile-by-toon/ZAELOT/30?request_flags=scr_mmgameloading"
	url := fmt.Sprint("http://localhost:", port, "/web-api/v2/", path)
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			body, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(string(body))
			}
		}
	}
	// myWindow.ShowAndRun()
}
