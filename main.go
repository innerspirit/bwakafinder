package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"

	"net/http"
	"os"
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
		fmt.Println(err)
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
	myWindow.SetContent(container.NewMax(container.NewMax(myTable), errorLabel))
	myWindow.Resize(fyne.NewSize(475, 400))

	dataChannel := make(chan [][]string)
	errorChannel := make(chan error)

	go func() {
		for {
			repdata := getReplayData(repPath + "\\LastReplay.rep")
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
				fullData = append(fullData, servData[1:]...)
			}
			if err != nil {
				errorChannel <- err
				time.Sleep(15 * time.Second)
				continue
			}
			dataChannel <- fullData
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

func getReplayData(fileName string) map[string]interface{} {
	cfg := repparser.Config{
		Commands: true,
		MapData:  true,
	}
	r, err := repparser.ParseFileConfig(fileName, cfg)
	if err != nil {
		fmt.Printf("Failed to parse replay: %v\n", err)
		os.Exit(1)
	}
	var destination = os.Stdout
	return compileReplayInfo(destination, r)
}

func grabPlayerInfo(player string, server string) ([][]string, error) {
	_, port, err := lib.GetProcessInfo(false)
	fmt.Println("sc port: ", port)
	path := "aurora-profile-by-toon/"
	params := fmt.Sprint("/", server, "?request_flags=scr_mmgameloading")
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
			body, readErr := ioutil.ReadAll(res.Body)
			if readErr != nil {
				return nil, readErr
			} else {
				var res MMGameLoadingRes
				json.Unmarshal(body, &res)
				response := [][]string{
					{"AKA", "Max MMR", "Rank"},
				}

				for _, stat := range res.MMStats {
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
		}
	}
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
