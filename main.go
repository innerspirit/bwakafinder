//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"fyne.io/fyne/v2"
)

var userHome, _ = os.UserHomeDir()
var repPath = userHome + "\\Documents\\StarCraft\\Maps\\Replays"

type Account struct {
	Account string `json:"account"`
}

type Settings struct {
	GatewayHistory []Account `json:"Gateway History"`
}

func main() {
	// --- load settings ---
	var settings Settings
	settingsFilePath := userHome + "\\Documents\\StarCraft\\CSettings.json"
	if f, err := os.Open(settingsFilePath); err == nil {
		defer f.Close()
		if b, err := ioutil.ReadAll(f); err == nil {
			_ = json.Unmarshal(b, &settings)
		}
	}

	// --- build account list ---
	var accounts []string
	for _, acct := range settings.GatewayHistory {
		accounts = append(accounts, acct.Account)
	}

	// --- channels for UI ↔ data goroutines ---
	dataCh := make(chan [][]string)
	errCh := make(chan error)

	// --- build & show the UI, then start data processing ---
	myWindow := NewUI(dataCh, errCh)

	// Save window position on close
	myWindow.SetCloseIntercept(func() {
		prefs := fyne.CurrentApp().Preferences()
		pos := myWindow.Position()
		prefs.SetFloat("window.x", float64(pos.X))
		prefs.SetFloat("window.y", float64(pos.Y))
		myWindow.Close()
	})

	// Set initial position
	prefs := fyne.CurrentApp().Preferences()
	x := prefs.FloatWithFallback("window.x", 100)
	y := prefs.FloatWithFallback("window.y", 100)
	myWindow.Move(fyne.NewPos(float32(x), float32(y)))

	StartDataProcessing(repPath, accounts, dataCh, errCh)
	myWindow.ShowAndRun()
}
