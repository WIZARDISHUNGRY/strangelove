package main

// https://ride.citibikenyc.com/system-data
// 40.688265,-73.9184594,21z

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"jonwillia.ms/strangelove/tui"
)

func main() {

	m := tui.NewModel()
	if err := tea.NewProgram(m).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
