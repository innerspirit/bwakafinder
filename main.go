//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
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
	var settings Settings
	settingsFilePath := userHome + "\\Documents\\StarCraft\\CSettings.json"
	if f, err := os.Open(settingsFilePath); err == nil {
		defer f.Close()
		if b, err := ioutil.ReadAll(f); err == nil {
			_ = json.Unmarshal(b, &settings)
		}
	}

	var accounts []string
	for _, acct := range settings.GatewayHistory {
		accounts = append(accounts, acct.Account)
	}

	dataCh := make(chan [][]string)
	errCh := make(chan error)

	myWindow := NewUI(dataCh, errCh)

	// main does not need its own close intercept; handled in UI builder

	StartDataProcessing(repPath, accounts, dataCh, errCh)
	myWindow.ShowAndRun()
}
