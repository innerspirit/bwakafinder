package main

import (
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

	dataChannel := make(chan [][]string)
	errorChannel := make(chan error)

	go func() {
		for {
			repdata := getReplayData(repPath + "\\LastReplay.rep")
			newData, err := grabPlayerInfo(repdata["winner"].(*screp.Player).Name)
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
