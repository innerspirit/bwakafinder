package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	screp "github.com/icza/screp/rep"
	"github.com/icza/screp/repparser"
	"github.com/innerspirit/getscprocess/lib"
)

// these mirror your old types from main.go:
type MMR struct {
	MMR  int    `json:"highest_rating"`
	Toon string `json:"toon"`
}
type MMGameLoadingRes struct {
	MMStats []MMR `json:"matchmaked_stats"`
}

// StartDataProcessing kicks off your file‐watch/parse loop
// and pushes table‐rows or errors into the provided channels.
func StartDataProcessing(repPath string, accounts []string, dataCh chan [][]string, errCh chan error) {
	// Define TableHeader using UI constants
	TableHeader := []string{HeaderAKA, HeaderMaxMMR, HeaderRank}

	go func() {
		var lastMod time.Time
		servers := []string{"10", "11", "20", "30"}

		for {
			repFile := repPath + "\\LastReplay.rep"

			if fi, err := os.Stat(repFile); err == nil {
				if fi.ModTime().Equal(lastMod) {
					time.Sleep(5 * time.Second)
					continue
				}
				lastMod = fi.ModTime()
			}

			// show spinner
			errCh <- nil // clear any previous error
			rData, err := getReplayData(repFile)
			if err != nil {
				dataCh <- [][]string{TableHeader}
				errCh <- err
				time.Sleep(5 * time.Second)
				continue
			}

			var full [][]string
			for _, serv := range servers {
				var (
					rows [][]string
					e    error
				)
				win := rData["winner"].(*screp.Player).Name
				los := rData["loser"].(*screp.Player).Name
				if stringInSlice(win, accounts) {
					rows, e = grabPlayerInfo(los, serv)
				} else {
					rows, e = grabPlayerInfo(win, serv)
				}
				if e != nil {
					errCh <- fmt.Errorf("server %s: %v", serv, e)
					continue
				}
				if len(rows) > 1 {
					full = append(full, rows[1:]...)
				}
			}

			if len(full) == 0 {
				dataCh <- [][]string{TableHeader}
				errCh <- fmt.Errorf("no valid data received from any server")
				time.Sleep(5 * time.Second)
				continue
			}

			// dedupe highest MMR
			uniq := map[string][]string{}
			for _, row := range full {
				if len(row) < 3 {
					continue
				}
				aka := row[0]
				m, _ := strconv.Atoi(row[1])
				if prev, ok := uniq[aka]; !ok {
					uniq[aka] = row
				} else {
					pm, _ := strconv.Atoi(prev[1])
					if m > pm {
						uniq[aka] = row
					}
				}
			}
			dedup := make([][]string, 0, len(uniq))
			for _, v := range uniq {
				dedup = append(dedup, v)
			}
			// prepend header
			header := TableHeader
			dataCh <- append([][]string{header}, dedup...)

			time.Sleep(5 * time.Second)
		}
	}()
}

func getReplayData(fileName string) (map[string]interface{}, error) {
	cfg := repparser.Config{Commands: true, MapData: true}
	r, err := repparser.ParseFileConfig(fileName, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse replay: %w", err)
	}
	return compileReplayInfo(r), nil
}

func grabPlayerInfo(player, server string) ([][]string, error) {
	// Define TableHeader using UI constants
	TableHeader := []string{HeaderAKA, HeaderMaxMMR, HeaderRank}

	pid, port, err := lib.GetProcessInfo(false)
	if err != nil {
		showErrorDialog(fmt.Sprintf("SC process detection failed (pid:%d port:%d): %v", pid, port, err))
		return nil, fmt.Errorf("SC process detection failed")
	}
	if port == -1 {
		showErrorDialog("StarCraft: Remastered is not running\nPlease launch the game first")
		return nil, fmt.Errorf("SC:R is not running or port not found")
	}

	url := fmt.Sprintf("http://localhost:%d/web-api/v2/aurora-profile-by-toon/%s/%s?request_flags=scr_mmgameloading",
		port, player, server)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var apiRes MMGameLoadingRes
	if err := json.Unmarshal(body, &apiRes); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	out := [][]string{TableHeader}
	for _, stat := range apiRes.MMStats {
		var rank string
		switch {
		case stat.MMR > 2471:
			rank = "S"
		case stat.MMR > 2015:
			rank = "A"
		case stat.MMR > 1698:
			rank = "B"
		case stat.MMR > 1549:
			rank = "C"
		case stat.MMR > 1427:
			rank = "D"
		case stat.MMR > 1137:
			rank = "E"
		default:
			rank = "F"
		}
		out = append(out, []string{stat.Toon, strconv.Itoa(stat.MMR), rank})
	}
	return out, nil
}

func compileReplayInfo(rep *screp.Replay) map[string]interface{} {
	rep.Compute()
	var winner, loser *screp.Player
	wid := rep.Computed.WinnerTeam
	hasWin := (wid != 0)
	for _, p := range rep.Header.Players {
		if p.Team == wid {
			winner = p
		} else {
			loser = p
		}
	}
	if !hasWin {
		winner = rep.Header.Players[0]
	}

	// map/engine logic omitted for brevity
	kl := rep.MapData.Name
	if kl == "" {
		kl = rep.Header.Map
	}
	duration := rep.Header.Duration().Truncate(time.Second).String()

	return map[string]interface{}{
		"winner":    winner,
		"loser":     loser,
		"len":       duration,
		"map":       kl,
		"hasWinner": hasWin,
	}
}

func stringInSlice(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
