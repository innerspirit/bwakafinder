package main

import (
	"image/color"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// NewUI builds the window, table, spinner & error label,
// sets up sorting, and starts the two goroutines that
// read from dataCh / errCh to update UI.
func NewUI(dataCh chan [][]string, errCh chan error) fyne.Window {
	myApp    := app.New()
	myWindow := myApp.NewWindow("BW Aka Finder")

	// optional icon
	if icon, err := fyne.LoadResourceFromPath("icon.ico"); err == nil {
		myWindow.SetIcon(icon)
	}

	// initial table data
	data := [][]string{{"AKA", "Max MMR", "Rank"}}

	// table setup
	table := widget.NewTable(
		func() (r, c int) { return len(data), len(data[0]) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("Content")
			lbl.Resize(fyne.NewSize(200, 20))
			return lbl
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(data[id.Row][id.Col])
			obj.(*widget.Label).Resize(fyne.NewSize(200, 20))
		},
	)
	table.SetColumnWidth(0, 150)
	table.SetColumnWidth(1, 150)
	table.SetColumnWidth(2, 150)

	// error label + spinner
	errLabel := canvas.NewText("", color.RGBA{255, 0, 0, 255})
	errLabel.TextStyle.Bold = true
	spinner := widget.NewProgressBarInfinite()
	spinner.Hide()

	topBar := container.NewVBox(errLabel, spinner)
	content := container.NewBorder(topBar, nil, nil, nil, table)
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(475, 400))

	// sorting state
	sortCol := 1
	sortAsc := false

	// click‐to‐sort handler
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row != 0 || len(data) <= 1 {
			return
		}
		if sortCol == id.Col {
			sortAsc = !sortAsc
		} else {
			sortCol = id.Col
			sortAsc = (id.Col != 1) // MMR default DESC, others ASC
		}
		rows := data[1:]
		sort.SliceStable(rows, func(i, j int) bool {
			vi, vj := rows[i][sortCol], rows[j][sortCol]
			if sortCol == 1 {
				mi, _ := strconv.Atoi(vi)
				mj, _ := strconv.Atoi(vj)
				if sortAsc {
					return mi < mj
				}
				return mi > mj
			}
			if sortAsc {
				return vi < vj
			}
			return vi > vj
		})
		table.Refresh()
	}

	// consume new data
	go func() {
		for newData := range dataCh {
			// recover from any panic
			func() {
				defer func() { recover() }()
				if len(newData) == 0 {
					return
				}
				// copy & sanitize
				local := make([][]string, len(newData))
				for i, row := range newData {
					safeRow := make([]string, 3)
					for j := 0; j < 3 && j < len(row); j++ {
						safeRow[j] = row[j]
					}
					local[i] = safeRow
				}
				data = local

				// resort if needed
				if len(data) > 1 {
					rows := data[1:]
					sort.SliceStable(rows, func(i, j int) bool {
						vi, vj := "", ""
						if sortCol < len(rows[i]) {
							vi = rows[i][sortCol]
						}
						if sortCol < len(rows[j]) {
							vj = rows[j][sortCol]
						}
						if sortCol == 1 {
							mi, _ := strconv.Atoi(vi)
							mj, _ := strconv.Atoi(vj)
							if sortAsc {
								return mi < mj
							}
							return mi > mj
						}
						if sortAsc {
							return vi < vj
						}
						return vi > vj
					})
				}

				table.Refresh()
				errLabel.Text = ""
				errLabel.Refresh()
			}()
		}
	}()

	// consume errors
	go func() {
		for e := range errCh {
			if e == nil {
				// clear previous error
				errLabel.Text = ""
				errLabel.Refresh()
				continue
			}
			txt := e.Error()
			if txt == "SC:R is not running or port not found" {
				txt = "StarCraft: Remastered is not running\nPlease launch the game first"
			}
			errLabel.Text = txt
			errLabel.Refresh()
		}
	}()

	return myWindow
}

// showErrorDialog is called by grabPlayerInfo to pop up a native notification.
func showErrorDialog(message string) {
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "Error",
		Content: message,
	})
} 