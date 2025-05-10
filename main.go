//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	screp "github.com/icza/screp/rep"
	"github.com/icza/screp/repparser"
	"github.com/innerspirit/getscprocess/lib"
)

var userHome, _ = os.UserHomeDir()
var repPath = userHome + "\\Documents\\StarCraft\\Maps\\Replays"

type Account struct {
	Account string `json:"account"`
}

type Settings struct {
	GatewayHistory []Account `json:"Gateway History"`
}

type MMR struct {
	MMR  int    `json:"highest_rating"`
	Toon string `json:"toon"`
}

type MMGameLoadingRes struct {
	MMStats []MMR `json:"matchmaked_stats"`
}

func main() {
	var settings Settings
	settingsFile, err := os.Open(userHome + "\\Documents\\StarCraft\\CSettings.json")
	if err != nil {
	} else {
		defer settingsFile.Close()
		byteValue, _ := ioutil.ReadAll(settingsFile)
		json.Unmarshal(byteValue, &settings)
	}
	var accounts []string
	for _, account := range settings.GatewayHistory {
		accounts = append(accounts, account.Account)
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("BW Aka Finder")

	// Set window icon
	icon, err := fyne.LoadResourceFromPath("icon.ico")
	if err == nil {
		myWindow.SetIcon(icon)
	}

	data := [][]string{
		{"AKA", "Max MMR", "Rank"},
	}

	myTable := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0]) //rows, columns
		},
		func() fyne.CanvasObject {
			item := widget.NewLabel("Content")
			item.Resize(fyne.Size{
				Width:  200,
				Height: 20,
			})
			return item
		},
		func(i widget.TableCellID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(data[i.Row][i.Col])
			item.(*widget.Label).Resize(fyne.Size{
				Width:  200,
				Height: 20,
			})
		},
	)

	errorLabel := canvas.NewText("", color.RGBA{255, 0, 0, 255}) // red color error label for displaying errors
	errorLabel.TextStyle.Bold = true
	myTable.SetColumnWidth(0, 150)
	myTable.SetColumnWidth(1, 150)
	myTable.SetColumnWidth(2, 150)

	// -------- Spinner shown while the table is loading --------
	spinner := widget.NewProgressBarInfinite()
	spinner.Hide() // hidden until we start a fresh load

	// stack   error-label  +  spinner  at the top of the window
	topBar  := container.NewVBox(errorLabel, spinner)
	content := container.NewBorder(topBar, nil, nil, nil, myTable)
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(475, 400))

	dataChannel := make(chan [][]string)
	errorChannel := make(chan error)

	// -------- Sorting helpers --------
	sortColumn := 1     // default "Max MMR"
	sortAsc := false    // default "highest first"

	// click-to-sort handler
	myTable.OnSelected = func(id widget.TableCellID) {
		if id.Row != 0 || len(data) <= 1 {      // only header row is clickable and need rows
			return
		}
		if sortColumn == id.Col {
			sortAsc = !sortAsc // toggle asc/desc
		} else {
			sortColumn = id.Col
			// default DESC for MMR column, ASC for the others
			if id.Col == 1 {
				sortAsc = false
			} else {
				sortAsc = true
			}
		}

		rows := data[1:]                       // work on a local slice
		sort.SliceStable(rows, func(i, j int) bool {
			vi, vj := rows[i][sortColumn], rows[j][sortColumn]
			// numeric comparison for MMR, string otherwise
			if sortColumn == 1 {
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
		myTable.Refresh()
	}

	// remember the last time the replay file changed
	var lastModTime time.Time

	go func() {
		for {
			repFile := repPath + "\\LastReplay.rep"

			// --- skip work if the file is unchanged ---
			if fi, fiErr := os.Stat(repFile); fiErr == nil {
				if fi.ModTime().Equal(lastModTime) {
					time.Sleep(5 * time.Second)
					continue
				}
				lastModTime = fi.ModTime()
			}

			// Show spinner while we process the new file
			spinner.Show()
			spinner.Refresh()

			repdata, repErr := getReplayData(repFile)
			if repErr != nil {
				// On error, just clear the table and try again next time
				dataChannel <- [][]string{{"AKA", "Max MMR", "Rank"}}
				errorChannel <- repErr
				spinner.Hide()
				spinner.Refresh()
				time.Sleep(5 * time.Second)
				continue
			}

			var fullData [][]string
			var servData [][]string
			var err error
			servers := []string{"10", "11", "20", "30"}

			for _, serv := range servers {
				if stringInSlice(repdata["winner"].(*screp.Player).Name, accounts) {
					servData, err = grabPlayerInfo(repdata["loser"].(*screp.Player).Name, serv)
				} else {
					servData, err = grabPlayerInfo(repdata["winner"].(*screp.Player).Name, serv)
				}

				// Skip if error but don't fail the whole operation
				if err != nil {
					errorChannel <- fmt.Errorf("server %s: %v", serv, err)
					continue
				}

				// Only add data if we have rows beyond header
				if len(servData) > 1 {
					fullData = append(fullData, servData[1:]...)
				}
			}

			// If we got no data from any server, send just the header
			if len(fullData) == 0 {
				dataChannel <- [][]string{{"AKA", "Max MMR", "Rank"}}
				errorChannel <- fmt.Errorf("no valid data received from any server")
				continue
			}

			// --- de-duplicate, keeping highest MMR for each AKA ---
			unique := map[string][]string{}
			for _, row := range fullData {
				if len(row) < 3 {               // ← ignore malformed rows
					continue
				}
				aka    := row[0]
				mmr, _ := strconv.Atoi(row[1])
				if prev, ok := unique[aka]; !ok {
					unique[aka] = row
				} else {
					prevMMR, _ := strconv.Atoi(prev[1])
					if mmr > prevMMR {
						unique[aka] = row
					}
				}
			}
			dedup := make([][]string, 0, len(unique))
			for _, row := range unique {
				dedup = append(dedup, row)
			}

			// prepend header
			header := []string{"AKA", "Max MMR", "Rank"}
			dataChannel <- append([][]string{header}, dedup...)

			spinner.Hide()
			spinner.Refresh()
			time.Sleep(5 * time.Second) // Reduced from 15 to 5 seconds for faster retries
		}
	}()

	go func() {
		for newData := range dataChannel {
			// Process data in a completely defensive way
			func() {
				// Catch and recover from any panic
				defer func() {
					if r := recover(); r != nil {
					}
				}()
				
				// Ensure we have at least one row (header)
				if len(newData) == 0 {
					return
				}
				
				// Make a safe copy of the data
				localData := make([][]string, len(newData))
				for i, row := range newData {
					// Make sure each row has 3 columns
					safeRow := make([]string, 3)
					for j := 0; j < len(row) && j < 3; j++ {
						safeRow[j] = row[j]
					}
					localData[i] = safeRow
				}
				
				// Update the global data safely
				data = localData
				
				// Only sort if we have data beyond the header
				if len(localData) > 1 {
					// Get the data rows only (skip header)
					dataRows := localData[1:]
					
					// Sort the data rows
					sort.SliceStable(dataRows, func(i, j int) bool {
						// Default comparison values
						vi, vj := "", ""
						
						// Safely get values for comparison
						if sortColumn < len(dataRows[i]) {
							vi = dataRows[i][sortColumn]
						}
						if sortColumn < len(dataRows[j]) {
							vj = dataRows[j][sortColumn]
						}
						
						// Use numeric comparison for MMR column
						if sortColumn == 1 {
							mi, _ := strconv.Atoi(vi)
							mj, _ := strconv.Atoi(vj)
							if sortAsc {
								return mi < mj
							}
							return mi > mj
						}
						
						// Use string comparison for other columns
						if sortAsc {
							return vi < vj
						}
						return vi > vj
					})
				}
				
				// Update the UI
				myTable.Refresh()
				errorLabel.Text = "" // clear error label on success
				errorLabel.Refresh()
			}()
		}
	}()

	go func() {
		for err := range errorChannel {
			// Make error messages more user-friendly
			errorText := err.Error()
			if errorText == "SC:R is not running or port not found" {
				errorText = "StarCraft: Remastered is not running\nPlease launch the game first"
			}
			errorLabel.Text = errorText
			errorLabel.Refresh()
		}
	}()
	myWindow.ShowAndRun()
}

func getReplayData(fileName string) (map[string]interface{}, error) {
	cfg := repparser.Config{
		Commands: true,
		MapData:  true,
	}
	r, err := repparser.ParseFileConfig(fileName, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse replay: %w", err)
	}
	return compileReplayInfo(os.Stdout, r), nil
}

func grabPlayerInfo(player string, server string) ([][]string, error) {
	pid, port, err := lib.GetProcessInfo(false)
	if err != nil {
		showErrorDialog(fmt.Sprintf("SC process detection failed (pid:%d port:%d): %v", pid, port, err))
		return nil, fmt.Errorf("SC process detection failed")
	}
	if port == -1 {
		showErrorDialog("StarCraft: Remastered is not running\nPlease launch the game first")
		return nil, fmt.Errorf("SC:R is not running")
	}

	path := "aurora-profile-by-toon/"
	params := fmt.Sprint("/", server, "?request_flags=scr_mmgameloading")
	url := fmt.Sprint("http://localhost:", port, "/web-api/v2/", path, player, params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpRes, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer httpRes.Body.Close()

	body, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var apiRes MMGameLoadingRes
	if err := json.Unmarshal(body, &apiRes); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	response := [][]string{
		{"AKA", "Max MMR", "Rank"},
	}

	for _, stat := range apiRes.MMStats {
		var rank string
		switch {
		case stat.MMR > 2471:
			rank = "S"
		case stat.MMR > 2015 && stat.MMR < 2470:
			rank = "A"
		case stat.MMR > 1698 && stat.MMR < 2014:
			rank = "B"
		case stat.MMR > 1549 && stat.MMR < 1697:
			rank = "C"
		case stat.MMR > 1427 && stat.MMR < 1548:
			rank = "D"
		case stat.MMR > 1137 && stat.MMR < 1426:
			rank = "E"
		default:
			rank = "F"
		}
		response = append(response, []string{stat.Toon, fmt.Sprint(stat.MMR), rank})
	}
	return response, nil
}

func compileReplayInfo(out *os.File, rep *screp.Replay) map[string]interface{} {
	rep.Compute()
	var winner, loser *screp.Player
	winnerID := rep.Computed.WinnerTeam
	hasWinner := (winnerID != 0)

	for _, p := range rep.Header.Players {
		if p.Team == winnerID {
			winner = p
		} else {
			loser = p
		}
	}
	if !hasWinner {
		winner = rep.Header.Players[0]
	}

	engine := rep.Header.Engine.ShortName
	if rep.Header.Version != "" {
		engine = engine + " " + rep.Header.Version
	}
	mapName := rep.MapData.Name
	if mapName == "" {
		mapName = rep.Header.Map // But revert to Header.Map if the latter is not available.
	}

	d := rep.Header.Duration()

	ctx := map[string]interface{}{
		"winner":    winner,
		"loser":     loser,
		"len":       d.Truncate(time.Second).String(),
		"map":       mapName,
		"hasWinner": hasWinner,
	}

	return ctx
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func showErrorDialog(message string) {
	// This will show a native Windows error dialog
	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "Error", 
		Content: message,
	})
}
