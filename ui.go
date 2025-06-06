package main

import (
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

// Add this constant at the top of the file
const (
	HeaderAKA    = "AKA"
	HeaderMaxMMR = "Max MMR"
	HeaderRank   = "Rank"
)

// FuturisticTheme creates a dark, futuristic theme
type FuturisticTheme struct{}

func (f FuturisticTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{10, 15, 25, 255} // Very dark blue-black
	case theme.ColorNameButton:
		return color.RGBA{20, 30, 50, 255} // Dark blue-gray
	case theme.ColorNameDisabledButton:
		return color.RGBA{15, 20, 30, 255}
	case theme.ColorNameForeground:
		return color.RGBA{0, 255, 200, 255} // Cyan-green
	case theme.ColorNameDisabled:
		return color.RGBA{100, 100, 100, 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{150, 150, 150, 255}
	case theme.ColorNamePressed:
		return color.RGBA{0, 200, 255, 255} // Bright cyan
	case theme.ColorNameSelection:
		return color.RGBA{0, 100, 150, 80} // Softer, lower-contrast selection
	case theme.ColorNameSeparator:
		return color.RGBA{0, 0, 0, 0} // Transparent
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
		return 16 // Larger base text
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

// NewUI builds the window, table, spinner & error label,
// sets up sorting, and starts the two goroutines that
// read from dataCh / errCh to update UI.
func NewUI(dataCh chan [][]string, errCh chan error) fyne.Window {
	myApp := app.NewWithID("com.innerspirit.bwakafinder")
	myApp.Settings().SetTheme(&FuturisticTheme{})

	myWindow := myApp.NewWindow("BW AKA FINDER")

	// optional icon
	if icon, err := fyne.LoadResourceFromPath("icon.ico"); err == nil {
		myWindow.SetIcon(icon)
	}

	// Load saved window position and size
	prefs := myApp.Preferences()
	winWidth := prefs.FloatWithFallback("window.width", 580)
	winHeight := prefs.FloatWithFallback("window.height", 500)
	myWindow.Resize(fyne.NewSize(float32(winWidth), float32(winHeight)))

	// Save window position and size on close
	myWindow.SetCloseIntercept(func() {
		prefs.SetFloat("window.width", float64(myWindow.Canvas().Size().Width))
		prefs.SetFloat("window.height", float64(myWindow.Canvas().Size().Height))
		myWindow.Close()
	})

	// initial table data
	data := [][]string{{HeaderAKA, HeaderMaxMMR, HeaderRank}}

	// Create a custom table with futuristic styling
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

			// Plain-ASCII header + data
			label.SetText(text)
			label.TextStyle = fyne.TextStyle{Bold: id.Row == 0}
			label.Resize(fyne.NewSize(220, 35))
		},
	)

	// Set larger column widths for the futuristic look
	table.SetColumnWidth(0, 180)
	table.SetColumnWidth(1, 160)
	table.SetColumnWidth(2, 120)

	// Create futuristic error label (always visible on errors)
	errLabel := canvas.NewText("", color.RGBA{255, 50, 50, 255}) // Bright red
	errLabel.TextStyle.Bold = true
	errLabel.TextSize = 16

	// Create an infinite progress-bar to use as a hovering "spinner"
	spinner := widget.NewProgressBarInfinite()
	spinner.Hide()

	// Create a transparent rectangle with only a stroke
	tableBorder := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	tableBorder.StrokeColor = color.RGBA{0, 150, 255, 200}
	tableBorder.StrokeWidth = 2

	// Layout with futuristic spacing — only show errors at top now
	topBar := container.NewVBox(errLabel)

	// Overlay the spinner centered on top of the table + border
	tableContainer := container.NewMax(
		tableBorder,
		table,
		container.NewCenter(spinner),
	)
	content := container.NewBorder(topBar, nil, nil, nil, tableContainer)

	myWindow.SetContent(content)

	// sorting state
	sortCol := 1
	sortAsc := false

	// click-to-sort handler
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
				// Only clear error when we have valid data
				if len(newData) > 1 {
					errLabel.Text = ""
					errLabel.Refresh()
				}
				// hide our "spinner" once fresh data arrives
				spinner.Hide()
			}()
		}
	}()

	// consume errors
	go func() {
		for e := range errCh {
			if e == nil {
				// show spinner while scanning
				spinner.Show()
				continue
			}
			txt := e.Error()
			if txt == "SC:R is not running or port not found" {
				txt = "STARCRAFT: REMASTERED NOT DETECTED\nPLEASE LAUNCH THE GAME FIRST"
			} else {
				txt = "ERROR: " + txt
			}
			errLabel.Text = txt
			errLabel.Refresh()
			spinner.Hide()
		}
	}()

	return myWindow
}

// showErrorDialog is called by grabPlayerInfo to pop up a native notification.
func showErrorDialog(message string) {
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "SYSTEM ALERT",
		Content: message,
	})
}
