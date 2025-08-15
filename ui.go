package main

import (
	"embed"
	"image/color"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.ico
var iconFS embed.FS

const (
	HeaderAKA    = "AKA"
	HeaderMaxMMR = "Max MMR"
	HeaderRank   = "Rank"
)

type FuturisticTheme struct{}

func (f FuturisticTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{10, 15, 25, 255}
	case theme.ColorNameButton:
		return color.RGBA{20, 30, 50, 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{15, 20, 30, 255}
	case theme.ColorNameForeground:
		return color.RGBA{0, 255, 200, 255}
	case theme.ColorNameDisabled:
		return color.RGBA{100, 100, 100, 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{150, 150, 150, 255}
	case theme.ColorNamePressed:
		return color.RGBA{0, 200, 255, 255}
	case theme.ColorNameSelection:
		return color.RGBA{0, 100, 150, 80}
	case theme.ColorNameSeparator:
		return color.RGBA{0, 0, 0, 0}
	case theme.ColorNameShadow:
		return color.RGBA{0, 0, 0, 100}
	case theme.ColorNameInputBackground:
		return color.RGBA{15, 25, 40, 255}
	case theme.ColorNameMenuBackground:
		return color.RGBA{20, 30, 50, 255}
	case theme.ColorNameOverlayBackground:
		return color.RGBA{0, 0, 0, 180}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (f FuturisticTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (f FuturisticTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (f FuturisticTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 16
	case theme.SizeNameCaptionText:
		return 14
	case theme.SizeNameHeadingText:
		return 24
	case theme.SizeNameSubHeadingText:
		return 20
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 24
	case theme.SizeNameScrollBar:
		return 16
	case theme.SizeNameScrollBarSmall:
		return 8
	default:
		return theme.DefaultTheme().Size(name)
	}
}

func NewUI(dataCh chan [][]string, errCh chan error) fyne.Window {
	myApp := app.NewWithID("com.innerspirit.bwakafinder")
	myApp.Settings().SetTheme(&FuturisticTheme{})

	myWindow := myApp.NewWindow("BW AKA FINDER")

	// Load embedded icon
	if iconData, err := iconFS.ReadFile("icon.ico"); err == nil {
		iconRes := fyne.NewStaticResource("icon.ico", iconData)
		myWindow.SetIcon(iconRes)
	}

	prefs := myApp.Preferences()
	winWidth := prefs.FloatWithFallback("window.width", 580)
	winHeight := prefs.FloatWithFallback("window.height", 500)

	myWindow.Resize(fyne.NewSize(float32(winWidth), float32(winHeight)))

	myWindow.SetCloseIntercept(func() {
		prefs.SetFloat("window.width", float64(myWindow.Canvas().Size().Width))
		prefs.SetFloat("window.height", float64(myWindow.Canvas().Size().Height))
		myWindow.Close()
	})

	data := [][]string{{HeaderAKA, HeaderMaxMMR, HeaderRank}}

	table := widget.NewTable(
		func() (r, c int) { return len(data), len(data[0]) },
		func() fyne.CanvasObject {
			lbl := widget.NewLabel("Content")
			lbl.Resize(fyne.NewSize(220, 35))
			lbl.TextStyle = fyne.TextStyle{Bold: true}
			return lbl
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			text := data[id.Row][id.Col]

			label.SetText(text)
			label.TextStyle = fyne.TextStyle{Bold: id.Row == 0}
			label.Resize(fyne.NewSize(220, 35))
		},
	)

	table.SetColumnWidth(0, 180)
	table.SetColumnWidth(1, 160)
	table.SetColumnWidth(2, 120)

	errLabel := canvas.NewText("", color.RGBA{255, 50, 50, 255})
	errLabel.TextStyle.Bold = true
	errLabel.TextSize = 16

	spinner := widget.NewProgressBarInfinite()
	spinner.Hide()

	tableBorder := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	tableBorder.StrokeColor = color.RGBA{0, 150, 255, 200}
	tableBorder.StrokeWidth = 2

	topBar := container.NewVBox(errLabel)

	tableContainer := container.NewMax(
		tableBorder,
		table,
		container.NewCenter(spinner),
	)
	content := container.NewBorder(topBar, nil, nil, nil, tableContainer)

	myWindow.SetContent(content)

	sortCol := 1
	sortAsc := false

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

	go func() {
		for newData := range dataCh {
			func() {
				defer func() { recover() }()
				if len(newData) == 0 {
					return
				}
				local := make([][]string, len(newData))
				for i, row := range newData {
					safeRow := make([]string, 3)
					for j := 0; j < 3 && j < len(row); j++ {
						safeRow[j] = row[j]
					}
					local[i] = safeRow
				}
				data = local

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

				fyne.Do(func() {
					table.Refresh()
					// Only clear error when we have valid data
					if len(newData) > 1 {
						errLabel.Text = ""
						errLabel.Refresh()
					}
					// hide our "spinner" once fresh data arrives
					spinner.Hide()
				})
			}()
		}
	}()

	go func() {
		for e := range errCh {
			fyne.Do(func() {
				if e == nil {
					// show spinner while scanning
					spinner.Show()
					return
				}
				txt := e.Error()
				if txt == "SC:R is not running or port not found" {
					txt = "STARCRAFT NOT DETECTED, PLEASE LAUNCH THE GAME FIRST"
				} else {
					txt = "ERROR: " + txt
				}
				errLabel.Text = txt
				errLabel.Refresh()
				spinner.Hide()
			})
		}
	}()

	return myWindow
}

func showErrorDialog(message string) {
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "SYSTEM ALERT",
		Content: message,
	})
}
