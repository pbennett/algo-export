package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func stateFile() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, fmt.Sprintf("algo-csv-state.json"))
}

// ExportState is the root type we use to persist state for the formats/accounts
// we exported.
// The state is tracking by format, then by account, and storing a 'state' instance.
type ExportState map[string]AccountExportState

type AccountExportState map[string]*state

type state struct {
	LastRound uint64
}

// LoadConfig hande
func LoadConfig() ExportState {
	var retState = ExportState{}
	configFile := stateFile()
	if !fileExist(configFile) {
		return retState
	}
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalln("reading config file:", configFile, "error:", err)
	}
	err = json.Unmarshal(configBytes, &retState)
	if err != nil {
		log.Fatalln("parsing config file:", configFile, "error:", err)
	}
	return retState
}

func (s ExportState) ForAccount(format string, account string) *state {
	if exportState, found := s[format]; found {
		if retState, found := exportState[account]; found {
			return retState
		}
		s[format][account] = &state{}
		return s[format][account]
	}
	s[format] = AccountExportState{account: &state{}}
	return s[format][account]
}

func (s ExportState) SaveConfig() {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Println("error mashalling address data:", err)
		os.Exit(1)
	}
	_ = ioutil.WriteFile(stateFile(), data, 0644)
}
